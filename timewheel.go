package timewheel

import (
    "container/list"
    "fmt"
    "time"
)

const (
    PrecisionSecond = iota
    PrecisionMillisecond = iota
    defaultChanCapacity = 1 << 8
)

var Nonblock bool

type Precision uint8

type Handler func(key string, data interface{}) error

type task struct {
    key      string
    data     interface{}
    cycle    int
    handler  Handler
    deadline time.Time
}

func NewTask(key string, data interface{}, deadline time.Time, handler Handler) *task {
    return &task{
        key:      key,
        data:     data,
        handler:  handler,
        deadline: deadline,
    }
}

type timeWheel struct {
    mask       int
    slots      []*list.List
    addCh      chan *task
    delCh      chan string
    stopCh     chan struct{}
    handler    Handler
    current    int
    interval   time.Duration
    errHandler func(err error)
}

func NewTimeWheel(precision Precision, handler Handler, errHandler func(err error)) *timeWheel {
    var (
        slot int
        interval time.Duration
    )

    switch precision {
    case PrecisionSecond:
        interval = time.Second
        slot = 60
    case PrecisionMillisecond:
        interval = time.Millisecond
        slot = 1000
    default:
        return nil
    }

    slots := make([]*list.List, slot)
    for k := range slots {
        slots[k] = list.New()
    }

    t := &timeWheel{
        interval:   interval,
        slots:      slots,
        mask:       slot,
        current:    0,
        addCh:      make(chan *task, defaultChanCapacity),
        delCh:      make(chan string, defaultChanCapacity),
        stopCh:     make(chan struct{}),
        handler:    handler,
        errHandler: errHandler,
    }

    go t.loop()

    return t
}

func (t *timeWheel)loop()  {
    ticker := time.NewTicker(t.interval)
    lp:
    for {
        select {
        case <-ticker.C:
            t.scan(t.current)
            t.current = (t.current + 1) % t.mask
        case k := <-t.addCh:
            t.add(k)
        case key := <-t.delCh:
            t.del(key)
        case <-t.stopCh:
            ticker.Stop()
            break lp
        }
    }
}

func (t *timeWheel)index(deadline time.Time) (cycle int, i int) {
    distance := deadline.UnixNano() - time.Now().UnixNano()
    if distance <= 0 {
        return 0, t.current
    }
    distance += int64(t.current) * int64(t.interval)
    cycle = int(distance / int64(t.interval) / int64(t.mask))
    i = int(distance / int64(t.interval) % int64(t.mask))
    return
}

func (t *timeWheel)add(k *task)  {
    var index int
    k.cycle, index = t.index(k.deadline)
    t.slots[index].PushBack(k)
}

func (t *timeWheel)del(key string)  {
    var e *list.Element
    for _, l := range t.slots {
        e = l.Front()
        for e != nil {
            k := e.Value.(*task)
            if k.key == key {
                evicted := e
                e = e.Next()
                l.Remove(evicted)
            } else {
                e = e.Next()
            }
        }
    }
}

func (t *timeWheel)scan(index int)  {
    l := t.slots[index]
    e := l.Front()
    for e != nil {
        k := e.Value.(*task)
        if k.cycle <= 0 {
            handler := k.handler
            if handler == nil {
                handler = t.handler
            }
            if handler == nil {
                continue
            } else if Nonblock {
                go t.Handle(k, handler, t.errHandler)
            } else {
                t.Handle(k, handler, t.errHandler)
            }
            evicted := e
            e = e.Next()
            l.Remove(evicted)
        } else {
            k.cycle --
            e = e.Next()
        }
    }
}

func (t *timeWheel)Handle(k *task, handler Handler, errHandler func(err error))  {
    defer func() {
        if err := recover(); err != nil && errHandler != nil {
            if _, ok := err.(error); ok {
                errHandler(err.(error))
            } else {
                errHandler(fmt.Errorf("%v", err))
            }
        }
    }()
    err := handler(k.key, k.data)
    if err != nil {
        panic(err)
    }
}

func (t *timeWheel)AddTask(k *task)  {
    t.addCh <- k
}

func (t *timeWheel)DelTask(key string)  {
    t.delCh <- key
}

func (t *timeWheel)Len() int {
    n := 0
    for _, l := range t.slots {
        n += l.Len()
    }
    return n
}

func (t *timeWheel)Stop() {
    t.stopCh <- struct{}{}
}
