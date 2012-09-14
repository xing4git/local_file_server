This is a little web app written with gloang.


Get:
go get github.com/xing4git/local_file_server


Before usage, you need prepare a properties file, it must fit syntax in java Properties. 

In the file, you should contains two configuration:

```
# this is the basic dir which you want to public
basedir=/home/xing
# this is the dir which will put all others upload files in
uploaddir=/home/xing/uploads
```

Says, your properties filename is my.conf, then:
```
local_file_server my.conf
```
Now, you can visit:

http://[your ip]:9090/upload

to upload file to uploaddir, or:

http://[your ip]:9090/local?path=[filepath]