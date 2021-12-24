# DedeCMS发版工具
DedeCMS发版工具。

```
1、执行：go run main.go
自动生成source文件夹，然后update文件夹自动生成当天的txt。例如：【20210101.file.txt】

2、执行：go run main.go
请输入本次所发版本的版本号: V5.7.***
自动生成public文件夹

2.1、如果需要手动放置sql文件
gb2312文件夹里面的sql文件格式为gb2312，文件名(年月日.sql)
utf-8文件夹里面的sql文件格式为utf-8，文件名(年月日.sql)
patch-xxx.zip压缩文件里面的sql文件格式为utf-8，文件名(年月日.sql)

3、更新后台版本
复制base-v57文件夹
在服务器上进入/public目录
粘贴后选择[全部应用、合并]文件夹
然后选择[全部应用、替换]文件
然后手动修改verinfo.txt，文件格式为gb2312

4、更新官网版本
复制base-v57文件夹
在服务器上进入/public目录
粘贴后选择[全部应用、合并]文件夹
然后选择[全部应用、替换]文件
然后手动修改verinfo.txt，文件格式为utf-8

5、在 https://github.com/dedecms/5.7/releases 生成一个Releases，V5.7.***
注意！！！字母V必须大写！！！
```
