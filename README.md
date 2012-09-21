This is a lightweight web app written with gloang.
```
Functionality:
1. upload file to your server.
2. view file in your server.
3. download file from your server.
```



Get:
go get github.com/xing4git/local_file_server


Before usage, you need prepare a properties file, it must fit syntax in java Properties. 

In the file, you should contains:
```
# this is the basic dir which you want to public
basedir=/home/xing
# this is the dir which will put all others upload files in
uploaddir=/home/xing/uploads
# this is the port listened by server:
port=9090
```

If your properties filename is my.conf, then start server:
```
local_file_server my.conf
```
Now, you can visit:

http://[your ip]:9090/upload

to upload file to uploaddir, or:

http://[your ip]:9090/local/[filepath]