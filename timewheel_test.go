package timewheel

import (
    "fmt"
    "math/rand"
    "sync"
    "testing"
    "time"
)

func TestNewTimeWheel(t *testing.T) {
    Nonblock = true
    tw := NewTimeWheel(PrecisionSecond, nil, nil)
    if tw == nil {
        t.Errorf("expected not nil")
    } else {
        n := 120
        g := sync.WaitGroup{}
        g.Add(n)
        now := time.Now()
        for i := 0; i < n; i++ {
            d := time.Now().Add(time.Second * time.Duration(i))
            key := fmt.Sprintf("task %d", i)
            k := NewTask(key, d, d, func(key string, data interface{}) error {
                fmt.Println(key, data.(time.Time).Sub(now))
                g.Done()
                return nil
            })
            tw.AddTask(k)
        }
        g.Wait()
        tw.Stop()
    }
}

func TestNewTimeWheel2(t *testing.T) {
    tw := NewTimeWheel(PrecisionSecond, nil, nil)
    if tw == nil {
        t.Errorf("expected not nil")
    } else {
        n := 2000
        g := sync.WaitGroup{}
        g.Add(n)
        now := time.Now()
        for i := 0; i < n; i++ {
            d := time.Now().Add(time.Millisecond * time.Duration(i))
            key := fmt.Sprintf("task %d", i)
            k := NewTask(key, d, d, func(key string, data interface{}) error {
                fmt.Println(key, data.(time.Time).Sub(now))
                g.Done()
                return nil
            })
            tw.AddTask(k)
        }
        g.Wait()
        tw.Stop()
    }
}

func TestNewTimeWheel3(t *testing.T) {
    tw := NewTimeWheel(PrecisionSecond, nil, func(err error) {
        fmt.Println(err)
    })
    if tw == nil {
        t.Errorf("expected not nil")
    } else {
        n := 2000
        g := sync.WaitGroup{}
        g.Add(n)
        now := time.Now()
        for i := 0; i < n; i++ {
            d := time.Now().Add(time.Millisecond * time.Duration(i))
            key := fmt.Sprintf("task %d", i)
            k := NewTask(key, d, d, func(key string, data interface{}) error {
                g.Done()
                return fmt.Errorf("%s %v", key, data.(time.Time).Sub(now))
            })
            tw.AddTask(k)
        }
        g.Wait()
        tw.Stop()
    }
}

func TestNewTimeWheel4(t *testing.T) {
    n := 300
    slice := rand.Perm(n)
    g := sync.WaitGroup{}
    g.Add(n)
    from := time.Now()
    tw := NewTimeWheel(PrecisionSecond, func(key string, data interface{}) error {
        return fmt.Errorf("%s %v", key, data.(time.Time).Sub(from))
    }, func(err error) {
        fmt.Println(err)
        g.Done()
    })
    if tw == nil {
        t.Errorf("expected not nil")
    } else {
        for _, i := range slice {
            d := time.Now().Add(time.Second * time.Duration(i))
            key := fmt.Sprintf("task %d", i)
            k := NewTask(key, d, d, nil)
            tw.AddTask(k)
        }
        g.Wait()
        tw.Stop()
    }
}

func TestNewTimeWheel5(t *testing.T) {
    n := 30000
    slice := rand.Perm(n)
    g := sync.WaitGroup{}
    g.Add(n)
    from := time.Now()
    tw := NewTimeWheel(PrecisionMillisecond, func(key string, data interface{}) error {
        return fmt.Errorf("%s %v", key, data.(time.Time).Sub(from))
    }, func(err error) {
        fmt.Println(err)
        g.Done()
    })
    if tw == nil {
        t.Errorf("expected not nil")
    } else {
        for _, i := range slice {
            d := time.Now().Add(time.Millisecond * time.Duration(i))
            key := fmt.Sprintf("task %d", i)
            k := NewTask(key, d, d, nil)
            tw.AddTask(k)
        }
        g.Wait()
        tw.Stop()
    }
}

func TestNewTimeWheel6(t *testing.T) {
    Nonblock = true
    n := 2000
    slice := rand.Perm(n)
    ch := make(chan error, n)
    from := time.Now()
    tw := NewTimeWheel(PrecisionMillisecond, func(key string, data interface{}) error {
        return fmt.Errorf("%s %v", key, data.(time.Time).Sub(from))
    }, func(err error) {
        ch <- err
    })
    if tw == nil {
        t.Errorf("expected not nil")
    } else {
        for _, i := range slice {
            d := time.Now().Add(time.Millisecond * time.Duration(i))
            key := fmt.Sprintf("task %d", i)
            k := NewTask(key, d, d, nil)
            tw.AddTask(k)
        }
        for err := range ch {
            fmt.Println(err)
        }
    }
}

func TestTimeWheel_DelTask(t *testing.T) {
    from := time.Now()
    tw := NewTimeWheel(PrecisionSecond, func(key string, data interface{}) error {
        return fmt.Errorf("%s %v", key, data.(time.Time).Sub(from))
    }, func(err error) {
        fmt.Println(err)
    })
    if tw == nil {
        t.Errorf("expected not nil")
    } else {
        slice := []int{2, 3}
        tasks := make([]*task, 0)
        for _, i := range slice {
            d := time.Now().Add(time.Second * time.Duration(i))
            key := fmt.Sprintf("task %d", i)
            k := NewTask(key, d, d, nil)
            tw.AddTask(k)
            tasks = append(tasks, k)
        }
        for _, task := range tasks {
            tw.DelTask(task.key)
        }
        time.Sleep(time.Minute)
    }
}

func TestNewTimeWheel7(t *testing.T) {
    globalHandler := func(key string, data interface{}) error {
        panic(fmt.Sprintf("key %s data %v panic", key, data))
    }
    errHandler := func(err error) {
        fmt.Println(err)
    }
    tw := NewTimeWheel(PrecisionSecond, globalHandler, errHandler)
    if tw != nil {
        task1 := NewTask("task1", "task1's data", time.Now().Add(time.Second * 3), nil)
        tw.add(task1)
    }
}