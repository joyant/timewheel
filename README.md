# timewheel

精度可选的时间轮算法，用`GoLang`实现。

## Example

一般用法。

```golang
import (
    "github.com/joyant/timewheel"
    "fmt"
    "time"
)

func main () {
    globalHandler := func(key string, data interface{}) error {
        panic(fmt.Sprintf("key %s data %v panic", key, data))
    }
    errHandler := func(err error) {
        fmt.Println(err)
    }
    tw := timewheel.NewTimeWheel(timewheel.PrecisionSecond, globalHandler, errHandler)
    if tw != nil {
        task1 := timewheel.NewTask("task1", "task1's data", time.Now().Add(time.Second * 3), nil)
        tw.add(task1)
    }
    quit := make(chan struct{})
    <-quit
}
```

每个任务都可以有自己的任务处理函数，任务自己的处理函数比绑定在`timewheel`上的任务处理函数优先级高。

```golang
import (
    "github.com/joyant/timewheel"
    "fmt"
    "time"
)

func main () {
    globalHandler := func(key string, data interface{}) error {
        panic(fmt.Sprintf("key %s data %v panic", key, data)) // 不会被执行
    }
    errHandler := func(err error) {
        fmt.Println(err) // 因为任务执行时并不会`panic`，所以不会被执行
    }
    task1Handler = func(key string, data interface{}) error {
        fmt.Printf("key %s data %v finished \n", key, data) // 会被执行
        return nil
    }
    tw := timewheel.NewTimeWheel(timewheel.PrecisionSecond, globalHandler, errHandler)
    if tw != nil {
        task1 := timewheel.NewTask("task1", "task1's data", time.Now().Add(time.Second * 3), task1Handler)
        tw.add(task1)
    }
    quit := make(chan struct{})
    <-quit
}
```

上面的例子精度都是秒，如果对精度要求比较高，可以用毫秒，只有这两种精度选择。

```golang
import (
    "github.com/joyant/timewheel"
    "fmt"
    "time"
)

func main () {
    task1Handler = func(key string, data interface{}) error {
        fmt.Printf("key %s data %v finished \n", key, data)
        return nil
    }
    tw := timewheel.NewTimeWheel(timewheel.PrecisionMillisecond, globalHandler, nil)
    if tw != nil {
        //任务将在3毫秒后执行
        task1 := timewheel.NewTask("task1", "task1's data", time.Now().Add(time.Millisecond * 3), task1Handler)
        tw.add(task1)
    }
    quit := make(chan struct{})
    <-quit
}
```

默认的，任务的执行是串行的(阻塞模式)，并没有跑在独立的协程中，如果任务执行比较耗时，需要作如下设置:

```golang
import (
    "github.com/joyant/timewheel"
    "fmt"
    "time"
)

func main () {
    timewheel.Nonblock = true // 这意味着每个任务执行都会新起一个协程

    task1Handler = func(key string, data interface{}) error {
        fmt.Printf("key %s data %v finished \n", key, data)
        return nil
    }
    tw := timewheel.NewTimeWheel(timewheel.PrecisionMillisecond, globalHandler, nil)
    if tw != nil {
        //任务将在3毫秒后执行
        task1 := timewheel.NewTask("task1", "task1's data", time.Now().Add(time.Millisecond * 3), task1Handler)
        tw.add(task1)
    }
    quit := make(chan struct{})
    <-quit
}
```

需要注意的是，做了`timewheel.Nonblock = true`设置后，`timewheel`并没有对协程的数量做限制，所以在任务较多的情况下，协程会无限制
增加，这并不是一个好的实践，如果任务较多，调用者不应该设置`timewheel.Nonblock = true`，而是应该把任务放在队列里，并且在接收任务的
时候要尽可能的快，比如这样：

```golang
import (
    "github.com/joyant/timewheel"
    "fmt"
    "time"
)

type MyTask struct {
    key string
    data interface{}
}

func main () {
    queue := make(chan MyTask, 1E4)
    task1Handler = func(key string, data interface{}) error {
        queue <- MyTask{ // 不要阻塞，要尽快的放入队列，不要在这里执行耗时任务，否则之后的任务会延迟
            key: key,
            data: data,
        }
        return nil
    }
    tw := timewheel.NewTimeWheel(timewheel.PrecisionMillisecond, globalHandler, nil)
    if tw != nil {
        task1 := timewheel.NewTask("task1", "task1's data", time.Now().Add(time.Millisecond * 3), task1Handler)
        tw.add(task1)
    }
    for _, q := range queue {
        fmt.Println(q.key, q.data)
    }
}
```

也可以删除任务和停止任务。

```golang
tw.DelTask("task1")
tw.Stop()
```

