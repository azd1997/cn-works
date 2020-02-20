// 测试流程：

// HTTPProxy
// 1. 启动代理服务器
// 2. 在设备上设置网络代理，代理地址填写为我们代理服务器监听的地址
//  	以我是用的Ubuntu18.04LTS为例，设置网络代理的方法是在
//			Settings -> Network -> VPN -> Network Proxy
// 			点击小齿轮图标进入子页面，勾选 Manual选项，并设置IP端口为
// 			代理服务器所监听的IP和Port，例如这里所使用的 0.0.0.0:8080
// 3. 使用浏览器访问一个网页, 例如http://kk2w.cc (必须是http协议连接)，
// 		有的网站支持https，如果不写上协议前缀，会自动使用https进行连接，这样代理就不起作用了
// 		回到http://kk2w.cc的例子，访问后，浏览器正常显示，代理服务器终端会输出
// 		请求信息
// 		2020/02/19 21:46:07 HTTPProxy 监听在 8080 端口
//		Received request: CONNECT prtas.videocc.net:443 127.0.0.1:54274
//		Received request: GET kk2w.cc 127.0.0.1:54276
//		Received request: GET kk2w.cc 127.0.0.1:54276
// 		...
// 		(当然，使用curl等工具或者自己写个简单的http客户端也是一样的)
//
// ps: 实际测试时，许多标https的网站的请求也会被代理服务器获取到，但是由于https会检查客户端的身份信息，
// 但现在的代理服务器无法处理这些，因此无法获得真正服务器的响应，客户端所有https的请求都会无法连接


// Proxy
// 步骤与HTTPProxy基本一致。
// 此外，在不进行系统级别或浏览器级别的网络代理配置时，可以使用curl工具手动指定代理
// curl -x http://127.0.0.1:8081/ -I https://www.baidu.com/
// curl -x http://127.0.0.1:8081/ -I http://www.baidu.com/
// -x 指定代理服务器， -I 指定目标服务器
package main
