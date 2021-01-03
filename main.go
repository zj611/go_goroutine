package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	// 构建一个通道
	ch := make(chan int)
	// 开启一个并发匿名函数
	go func() {
		// 从3循环到0
		for i := 3; i >= 0; i-- {
			// 发送3到0之间的数值
			ch <- i
			// 每次发送完时等待
			time.Sleep(time.Second)
		}
	}()
	// 遍历接收通道数据
	for data := range ch {
		// 打印通道数据
		fmt.Println(data)
		// 当遇到数据0时, 退出接收循环
		if data == 0 {
			break
		}
	}
}

//-----------------------------------------------
//数据接收服务主协程同子协程同步变量
var wg sync.WaitGroup

func run(i int) {
	fmt.Println("start 任务ID：", i)
	time.Sleep(time.Second * 1)
	wg.Done() // 每个goroutine运行完毕后就释放等待组的计数器
}

func main() {
	countThread := 2 //runtime.NumCPU()
	for i := 0; i < countThread; i++ {
		go run(i)
	}
	wg.Add(countThread) // 需要开启的goroutine等待组的计数器

	//等待所有的任务都释放
	wg.Wait()
	fmt.Println("任务全部结束,退出")
}

//-----------------------------------------------
// 通过使用chan + select多路复用的方式
// 来通知协程结束
func run(stop chan bool) {
	for {
		select {
		case <-stop:
			fmt.Println("任务1结束")
			return
		default:
			fmt.Println("任务正在进行中")
			time.Sleep(time.Second * 2)
		}
	}
}

func main() {
	stop := make(chan bool)
	go run(stop) // 开启goroutine

	time.Sleep(time.Second * 5)
	fmt.Println("准备结束任务...")
	stop <- true
	time.Sleep(time.Second * 3)
	return
}

//-----------------------------------------------
// context
//context常用的使用场景：
//1) 一个请求对应多个goroutine之间的数据交互
//2) 超时控制
//3) 上下文控制
//
func run(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("任务%v结束退出\n", id)
			return
		default:
			fmt.Printf("任务%v正在运行中\n", id)
			time.Sleep(time.Second * 2)
		}
	}
}
func main() {
	//管理启动的协程
	ctx, cancel := context.WithCancel(context.Background())
	//fmt.Println(reflect.TypeOf(cancel))
	// 开启多个goroutine，传入ctx
	go run(ctx, 1)
	go run(ctx, 2)

	// 运行一段时间后停止
	time.Sleep(time.Second * 5)
	fmt.Println("停止任务...")
	// 使用context的cancel函数停止goroutine
	cancel() //context.CancelFunc

	// 为了检测监控过是否停止，如果没有监控输出，表示停止
	time.Sleep(time.Second * 3)
	return
}

//context.Background() 返回一个空的 Context，
//这个空的 Context 一般用于整个 Context 树的根节点。
//然后使用 context.WithCancel(parent) 函数，
//创建一个可取消的子 Context，然后当作参数传给 goroutine 使用，
//这样就可以使用这个子 Context 跟踪这个 goroutine。

//-----------------------------------------------
// context 超时控制

func run(ctx context.Context, duration time.Duration, id int, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("协程 %d 退出\n", id)
			wg.Done()
			return
		case <-time.After(duration): // 注意，此处非超时，只是打印定时，超时是5s
			fmt.Printf("协程 %d 正在运行\n", id)
		}
	}
}

func main() {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go run(ctx, 1*time.Second, i, wg)
	}
	wg.Wait()
}

//-----------------------------------------------
// context 传参数
var key string = "name"

func run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("任务%v结束退出\n", ctx.Value(key))
			return
		default:
			fmt.Printf("任务%v正在运行中\n", ctx.Value(key))
			time.Sleep(time.Second * 2)
		}
	}
}

func main() {
	//管理启动的协程
	ctx, cancel := context.WithCancel(context.Background())
	// 给ctx绑定键值，传递给goroutine
	valuectx := context.WithValue(ctx, key, "【监控1】")

	// 开启goroutine，传入ctx
	go run(valuectx)

	// 运行一段时间后停止
	time.Sleep(time.Second * 10)
	fmt.Println("停止任务")
	cancel() // 使用context的cancel函数停止goroutine

	// 为了检测监控过是否停止，如果没有监控输出，表示停止
	time.Sleep(time.Second * 3)
}

// 总结
//1) 不要把 Context 放在结构体中，要以参数的方式传递
//2) 以 Context 作为参数的函数方法，应该把 Context 作为第一个参数，放在第一位
//3) 给一个函数方法传递 Context 的时候，不要传递 nil，如果不知道传递什么，就使用 context.TODO
//4) Context 的 Value 相关方法应该传递必须的数据，不要什么数据都使用这个传递
//5) Context 是线程安全的，可以放心的在多个 goroutine 中传递
