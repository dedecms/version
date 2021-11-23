# DedeCMS发版工具
DedeCMS发版工具。

```
1、执行：go run main.go
自动生成source文件夹，然后update文件夹自动生成当天的txt。例如：【20210101.file.txt】

2、执行：go run main.go
请输入本次所发版本的版本号: V5.7.***
自动生成public文件夹

3、更新后台版本
复制base-v57文件夹
在服务器上进入/public目录
粘贴后选择合并文件夹，
然后选择覆盖文件
然后手动修改verinfo.txt

4、更新官网版本
复制base-v57文件夹
在服务器上进入/public目录
粘贴后选择合并文件夹，
然后选择覆盖文件
然后手动修改verinfo.txt

5、在 https://github.com/dedecms/5.7/releases 生成一个Releases，V5.7.***
注意！！！字母V必须大写！！！
```
