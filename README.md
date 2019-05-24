# Bing Search Tool / 必应搜索工具
---------

 1. support keyword,hostname,ip address query / 支持关键字,域名,IP地址查询
 2. suport Class C IP address query / 支持C段查询
 3. use goroutine / 支持多线程查询
 4. support all platform / 全平台支持
 
---------
Usage/使用说明:`bing [-p proxy-url] [-t C/S] [-h hostOrip] [-k keyword] [-s stop-count] [-w worker thread]`  

Example/使用方法:`./bing -h www.google.com -t C -w 5`  
results will be write to `result.html` / 文件结果将会保存到`result.html`  
Test on `Macos Mojave` / `Windows 10`  

Options:
  -h string  
    	search host or ip address / 填域名或者IP地址   
  -k string  
    	force use keyword search / 强制使用关键字搜索  
  -p string  
    	proxy url,like http://127.0.0.1:1087 / 代理地址  
  -s int  
    	if result-count > x will stop, default 0 is not limited / 单个ip搜索结果大于该数则停止继续获取  
  -t string  
    	search type, value is C or S(single host) (default "S") / 搜索类型,S为单个IP/host, C表示使用C段搜索  
  -w int  
    	worker count (default 10) / 线程数量,调太大可能被墙  
