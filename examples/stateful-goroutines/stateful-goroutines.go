// В предыдущем примере мы использовали явную блокировку
// мьютексами для синхронизации доступа к общему состоянию
// между несколькими горутинами. Другой вариант заключается
// в использовании встроенных функций синхронизации горутин
// и каналов для достижения того же результата. Этот
// подход, основанный на каналах, соответствует идеям Go
// о разделении памяти через взаимодействие и того, что
// каждый фрагмент данных принадлежит ровно одной горутине.

package main

import (
    "fmt"
    "math/rand"
    "sync/atomic"
    "time"
)

// В этом примере наше состояние принадлежит одной
// горутине. Это гарантирует, что данные никогда не будут
// скомпрометированы конкурентным доступом. Для того,
// чтобы читать или изменять состояние, другие горутины
// будут отправлять сообшения владеющей горутине и
// получать соответствующие ответы. `Структуры` `readOp`
// и `writeOp` инкапсулируют эти запросы и способы ответа
// владеющей горутины.
type readOp struct {
    key  int
    resp chan int
}
type writeOp struct {
    key  int
    val  int
    resp chan bool
}

func main() {

    // Как и прежде мы будем считать, сколько мы произвели операций
    var ops int64 = 0

    // Каналя `reads` и `writes` будут использованы другими
    // горутинами для чтения и записи запросов соответственно.
    reads := make(chan *readOp)
    writes := make(chan *writeOp)

    // Здесь горутина, которой принадлежит `state`, который
    // Here is the goroutine that owns the `state`, which
    // is a map as in the previous example but now private
    // to the stateful goroutine. This goroutine repeatedly
    // selects on the `reads` and `writes` channels,
    // responding to requests as they arrive. A response
    // is executed by first performing the requested
    // operation and then sending a value on the response
    // channel `resp` to indicate success (and the desired
    // value in the case of `reads`).
    go func() {
        var state = make(map[int]int)
        for {
            select {
            case read := <-reads:
                read.resp <- state[read.key]
            case write := <-writes:
                state[write.key] = write.val
                write.resp <- true
            }
        }
    }()

    // This starts 100 goroutines to issue reads to the
    // state-owning goroutine via the `reads` channel.
    // Each read requires constructing a `readOp`, sending
    // it over the `reads` channel, and the receiving the
    // result over the provided `resp` channel.
    for r := 0; r < 100; r++ {
        go func() {
            for {
                read := &readOp{
                    key:  rand.Intn(5),
                    resp: make(chan int)}
                reads <- read
                <-read.resp
                atomic.AddInt64(&ops, 1)
            }
        }()
    }

    // We start 10 writes as well, using a similar
    // approach.
    for w := 0; w < 10; w++ {
        go func() {
            for {
                write := &writeOp{
                    key:  rand.Intn(5),
                    val:  rand.Intn(100),
                    resp: make(chan bool)}
                writes <- write
                <-write.resp
                atomic.AddInt64(&ops, 1)
            }
        }()
    }

    // Let the goroutines work for a second.
    time.Sleep(time.Second)

    // Finally, capture and report the `ops` count.
    opsFinal := atomic.LoadInt64(&ops)
    fmt.Println("ops:", opsFinal)
}
