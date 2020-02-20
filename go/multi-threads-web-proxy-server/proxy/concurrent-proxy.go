package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

//TODO: 存在问题，暂无法解决，留坑待补。按照报错信息，似乎是Hijack()必须在ServeHTTP结束之前调用
// 也就是说，不能让ServeHTTP结束，而让真正的处理在后面继续处理，这也就是要想并发，只能提升ServeHTTP的并发
// 当然可以再设置一个通道用来通知ServeHTTP退出，但是这样一点并发的效果都没有
// 这样的话，所谓的并发HTTP代理


func StartConcurrentProxy(addr, port string, anony, debug bool, maxGoroutines int) {
	proxy := NewConcurrentProxy(addr, port, anony, debug, maxGoroutines)
	defer proxy.Shutdown()


	log.Printf("Proxy is runing on %s:%s \n", proxy.addr, proxy.port)
	log.Fatalln(http.ListenAndServe(proxy.addr + ":" + proxy.port, proxy))
}



// 支持并发的代理
// 所谓支持并发，就是能同时处理多个连接。但是又不能无限制开很多个连接，因此需要一个连接池结构
// 可以理解为一个连接池代表了限制使用的一批资源，客户端每一个访问占用一个资源，当没有资源可用时，让其排队
// 直到有资源被释放出来可用.
// 由于HTTP是短连接，就是一次TCP连接只传一次数据就关掉(当然现在http也有长连接了，但这些不细究)
// 假如同时允许处理10个客户端连接conn，那么就相应的有个conn池，容量为10
// 另一个层面讲，加入最多要能同时支持10个连接，就得有10个goroutine（或者其他语言的线程/协程）
// 其实真正严格来说，前面说的分为两种：工作池和连接池，一个是处理事务，一个是资源，不太一样，但现在对这些还不太清楚
// 只能混为一谈，姑且认为连接池主要用于数据库的多连接、网络的多连接
// 这里呢，实现一个worker线程池，Proxy每次接收到一个请求，它不是自己处理，而是丢给workers去做
// worker池相关的代码，在当前文件的后半部分
type ConcurrentProxy struct {
	addr string			// 监听地址
	port string			// 监听端口
	isAnonymous bool	// 是否高匿名模式
	debug bool			// 调试模式

	// 自身池化
	works chan Work
	wg sync.WaitGroup

	// 引用池的做法被废弃
	//WorkerPool *Pool	// woker pool 工作池,// 工作池最大并发数，建议设置为CPU核数
}

type Work struct {
	w *http.ResponseWriter
	r *http.Request
}

func NewWork(w *http.ResponseWriter, r *http.Request) Work {
	return Work{w, r}
}

func NewConcurrentProxy(addr, port string, anony, debug bool, maxGoroutines int) *ConcurrentProxy {
	proxy := &ConcurrentProxy{
		addr:        addr,
		port:        port,
		isAnonymous: anony,
		debug:       debug,
		works:make(chan Work),	// 无缓冲通道
	}


	// 池化
	proxy.wg.Add(maxGoroutines)
	for i:=0; i<maxGoroutines; i++ {	// 这种做法是一直开着 maxGouroutine个协程
		// 开10个worker
		go func(i int) {
			// 对chan进行for-range会一直等channel传数据过来，一直等一直等，等到一个继续等下一个，不会退出
			// 一旦接收到新的work，这里边的代码才会继续执行，否则会继续等待
			for work := range proxy.works {
				log.Printf("worker [%d] 接收到新任务\n", i)
				// 真正处理
				proxy.serve(*work.w, work.r)	// 工作完了
				log.Printf("worker [%d] 执行完任务\n", i)
			}
			proxy.wg.Done()		// 协程结束
		}(i)
	}

	return proxy
}

// 实现http.Handler接口
func (proxy *ConcurrentProxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// 调试模式下，打印接收到的请求
	if proxy.debug {
		log.Printf("Received request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)
	}

	// 将请求塞入工作池
	proxy.works <- NewWork(&rw, r)

	// 等待结束
	select{}
}

func (proxy *ConcurrentProxy) serve(rw http.ResponseWriter, r *http.Request) {

	// 如果方法不是CONNECT就当HTTP处理，是的话，进行HTTPS直连
	if r.Method != http.MethodConnect {
		proxy.serveHttp(rw, r)
	} else {
		// 直通模式不作任何处理
		proxy.serveHttps(rw, r)
	}
}

func (proxy *ConcurrentProxy) serveHttp(w http.ResponseWriter, r *http.Request) {
	// 并发版本不能使用默认的Transport(单例模式)
	// transport := http.DefaultTransport
	var transport http.RoundTripper = &http.Transport{
		//Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 1. 复制原请求，并加以修改
	outReq := new(http.Request)
	*outReq = *r	// 复制内容

	if !proxy.isAnonymous {		// 不匿名的话，把来路的节点地址给加在"X-Forwarded-For"字段
		if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			// 查看请求头中"X-Forwarded-For"字段是否有值
			// 这个字段会存转发路径上所有节点地址，而且每一次都是当前节点去填写上一个节点的地址
			if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
				// 拼接路径上所有节点地址
				clientIP = strings.Join(prior, ",") + "," + clientIP
			}
			// 重新填入
			outReq.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	// 2. 将修改后的请求转发出去，等待服务器回应
	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		// 出错，那么说明这台代理服务器无法连通到，那么在w的状态头中写上BadGateway
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()

	// 3. 回写http头。 将真正服务器的状态头给写到w的状态头去(就是写回给客户端了)
	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	// 回写状态码
	w.WriteHeader(resp.StatusCode)
	// 回写响应内容
	_, _ = io.Copy(w, resp.Body)
}

func (proxy *ConcurrentProxy) serveHttps(w http.ResponseWriter, r *http.Request) {
	// 取出host
	host := r.URL.Host
	hij, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("HTTP Server does not support hijacking\n")
	}

	client, _, err := hij.Hijack()
	if err != nil {
		log.Fatalln("Hijack: ", err)
	}

	// 连接远程
	server, err := net.Dial("tcp", host)
	if err != nil {
		log.Fatalln("Dial host: ", err)
	}
	client.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	// 直通双向复制
	go io.Copy(server, client)
	go io.Copy(client, server)
}


func (proxy *ConcurrentProxy) Shutdown() {
	// 不再接收新任务
	close(proxy.works)
	proxy.wg.Wait()		// 等待所有worker协程退出
}



// ================== worker pool =================
// NOTICE: !以下实现的Pool作废，HTTP代理实现的过程中不允许将w,r
// 如此传递，实际跑起来有bug，最后选择将ConcurrentProxy自身池化
// 这样可以避免调用者的变化
// 更细节的解释暂时无法给出
//
// 更新：将池化转移到ConcurrentProxy同样报了相同的错，说明问题不在于池化交给谁
//


// 一个Pool代表了指定数量的并发Worker(尽管这个Worker并没有显示定义)
// 这个Pool也没有做过多的任务管理，不涉及任何的任务的优先度等问题
type Pool struct {
	proxy *ConcurrentProxy
	works chan ProxyWork
	wg sync.WaitGroup	// 使用wg来限制最大的协程数，当然也可以用其他的方式，比如计数，这里不表
}

// 每一个work都是一个函数，处理一些事情，由于需要传递r和w两个参数
// 所以包装成一个结构体比较好，再上层抽象一个接口，这样子通用性更好
// 但是这里只需要serve一个函数，所以真正需要的是把新的参数给传过去
//
// 由于ConcurrentProxy的方法和属性都在其那里，那么这里可以选择每一个
// 1)ProxyWork都引用一个ConcurrentProxy以使用其内部值及方法，
// 2)也可以将ConcurrentProxy的方法移到Work这来
// 3)将所有需要的参数都拷贝一份到ProxyWork来
// 4)把proxy的引用放到Pool中去。这样相当于ProxyWork只起传参作用
// 这里选择第四种做法
type ProxyWork struct {
	w http.ResponseWriter
	r *http.Request
}

func NewProxyWork(w http.ResponseWriter, r *http.Request) ProxyWork {
	return ProxyWork{w, r}
}

func NewPool(maxGoroutines int, proxy *ConcurrentProxy) *Pool {
	p := Pool{proxy:proxy, works:make(chan ProxyWork)}
	p.wg.Add(maxGoroutines)
	for i:=0; i<maxGoroutines; i++ {	// 这种做法是一直开着 maxGouroutine个协程
		// 开10个worker
		go func(i int) {
			// 对chan进行for-range会一直等channel传数据过来，一直等一直等，等到一个继续等下一个，不会退出
			// 一旦接收到新的work，这里边的代码才会继续执行，否则会继续等待
			for work := range p.works {
				log.Printf("worker [%d] 接收到新任务\n", i)
				p.proxy.serve(work.w, work.r)	// 工作完了
				log.Printf("worker [%d] 执行完任务\n", i)
			}
			p.wg.Done()		// 协程结束
		}(i)
	}

	return &p
}

func (p *Pool) Do(work ProxyWork) {
	// 将任务提交到works通道，由于works是无缓冲通道，
	// 必须等某一个goroutine接收了这个work，当前的Do()才会退出
	// 而退出代表work开始被真正的worker执行
	p.works <- work
}

func (p *Pool) Shutdown() {
	// 不再接收新任务
	close(p.works)
	p.wg.Wait()		// 等待所有worker协程退出
}

