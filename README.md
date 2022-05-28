# golang-multiparttest
Debugging and testing for multipart uploads in golang

This repository checks the behavior of https://pkg.go.dev/net/http#Request.ParseMultipartForm

There was a observation in my upload module (caddyv2-upload)[https://github.com/git001/caddyv2-upload]
that the memory usage is higher then expected `maxMemory` parameter from 
`ParseMultipartForm`.  

I have created this repository to see how the `maxMemory` parameter and
`ParseMultipartForm` behaves.

## create test files

```shell
dd if=/dev/zero of=1g.img bs=1024M count=1
dd if=/dev/zero of=10g.img bs=1024M count=10
```

## run main.go

The program expects 2 parameters.

### maxMemory size
```
-b 500M or -maxformbuffer 500M
```

### maximum file size
```
-s 1G   or -maxfilesize 1G
```

## Observations

### Temporary files

At the upload time will a temporary file created in the OS specific temporary
directory with this function (CreateTemp)[https://pkg.go.dev/os#CreateTemp] when
the file size is bigger then `maxMemory`

You can see this in this code block https://cs.opensource.google/go/go/+/refs/tags/go1.18.2:src/mime/multipart/formdata.go;l=91;drc=7791e934c882fd103357448aee0fd577b20013ce

```shell
ls -larth /tmp/multipart-*
```

### maxMemory

The parameter `maxMemory` is used to decides if a temporary file is created and
the memory usage of the upload Program.  
What I have seen is that the memory usage of the program is 2 times higher then
the  `maxMemory` setting.

```
go run main.go -b 500M -s 15G # > RES 1G
go run main.go -b 50M -s 15G  # > RES 100M
```

As I'm not very familiar with the `pprof` tool I can't dig deeper into the topic.
My conclusion is that the parameter `maxMemory` is a quite important one which
handles the balance between memory usage and file system usage.

## Tests

I have used 4 Shells for the test setup

### shell 1 the program
```
go run main.go -b 1M -s 15G
```

### shell 2 the temp directory watcher

There is a temporary file created with the size of the uploaded file.

```
ls -larth /tmp/multipart-*
-rw------- 1 alex alex 9,3G Mai 28 23:51 /tmp/multipart-933947771
```

### shell 3 the upload
```
curl -v --form myFile=@10g.img http://localhost:5050/
```

### shell 4 top

As you an see the 1M of the buffer limits also the memory of the upload program

```
# top -E k -b -d 1 -p 1659822 -p 1659698
top - 23:57:50 up 9 days, 12:50,  1 user,  load average: 2,12, 2,15, 2,20
Tasks:   2 total,   0 running,   2 sleeping,   0 stopped,   0 zombie
%Cpu(s):  7,3 us,  4,5 sy,  0,0 ni, 87,7 id,  0,0 wa,  0,0 hi,  0,6 si,  0,0 st
KiB Mem : 65542988 total, 17710860 free, 32479204 used, 15352924 buff/cache
KiB Swap:  8388604 total,   811380 free,  7577224 used. 32065292 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1659822 alex      20   0 1077276   5208   4268 S   0,0   0,0   0:00.00 main
1659698 alex      20   0 1897532  18040  10036 S   0,0   0,0   0:00.24 go

top - 23:57:51 up 9 days, 12:50,  1 user,  load average: 2,12, 2,15, 2,20
Tasks:   2 total,   0 running,   2 sleeping,   0 stopped,   0 zombie
%Cpu(s):  8,4 us,  3,7 sy,  0,0 ni, 87,8 id,  0,1 wa,  0,0 hi,  0,0 si,  0,0 st
KiB Mem : 65542988 total, 17710860 free, 32479204 used, 15352924 buff/cache
KiB Swap:  8388604 total,   811380 free,  7577224 used. 32065292 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1659822 alex      20   0 1077276   5208   4268 S   0,0   0,0   0:00.00 main
1659698 alex      20   0 1897532  18040  10036 S   0,0   0,0   0:00.24 go

top - 23:57:52 up 9 days, 12:50,  1 user,  load average: 2,12, 2,15, 2,20
Tasks:   2 total,   0 running,   2 sleeping,   0 stopped,   0 zombie
%Cpu(s):  9,6 us,  4,8 sy,  0,0 ni, 85,6 id,  0,0 wa,  0,0 hi,  0,0 si,  0,0 st
KiB Mem : 65542988 total, 17710860 free, 32479204 used, 15352924 buff/cache
KiB Swap:  8388604 total,   811380 free,  7577224 used. 32065292 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1659822 alex      20   0 1077276   7312   4268 S  19,0   0,0   0:00.19 main
1659698 alex      20   0 1897532  18040  10036 S   0,0   0,0   0:00.24 go

....

top - 23:58:25 up 9 days, 12:51,  1 user,  load average: 2,89, 2,35, 2,27
Tasks:   2 total,   0 running,   2 sleeping,   0 stopped,   0 zombie
%Cpu(s): 10,2 us,  5,0 sy,  0,0 ni, 84,6 id,  0,0 wa,  0,0 hi,  0,1 si,  0,0 st
KiB Mem : 65542988 total,  9207616 free, 32462172 used, 23873200 buff/cache
KiB Swap:  8388604 total,   808820 free,  7579784 used. 32074440 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1659822 alex      20   0 1077532   8456   5260 S   0,0   0,0   0:31.43 main
1659698 alex      20   0 1897532  18040  10036 S   0,0   0,0   0:00.24 go^C
```

### shell 1 the program

Now let's run the upload file with the same buffer size as the max file size
```
go run main.go -b 15G -s 15G
```

### shell 2 the temp directory watcher

There is no temporary file created.

```
ls -larth /tmp/multipart-*
ls: cannot access '/tmp/multipart-*': No such file or directory
```

### shell 4 top

As you an see the 15G of the buffer increases the memory usage for at least 2 times
of the set buffer size. The Buffers are not free after the upload

  RES   
 26,9g

```
top -E k -b -d 1 -p 1660918 -p 1660800
top - 00:04:24 up 9 days, 12:57,  1 user,  load average: 2,23, 2,64, 2,47
Tasks:   2 total,   0 running,   2 sleeping,   0 stopped,   0 zombie
%Cpu(s): 17,2 us,  6,3 sy,  0,0 ni, 76,4 id,  0,0 wa,  0,0 hi,  0,0 si,  0,0 st
KiB Mem : 65542988 total,  9398500 free, 32590608 used, 23553880 buff/cache
KiB Swap:  8388604 total,   837748 free,  7550856 used. 31957280 avail Mem 

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660800 alex      20   0 1824120  16736   9928 S   6,7   0,0   0:00.29 go
1660918 alex      20   0 1077532   5416   4496 S   0,0   0,0   0:00.00 main

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0 1077532   5416   4496 S   0,0   0,0   0:00.00 main
1660800 alex      20   0 1824120  16736   9928 S   0,0   0,0   0:00.29 go

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0 2677752 961736   5424 S  66,0   1,5   0:00.66 main
1660800 alex      20   0 1824120  16736   9928 S   0,0   0,0   0:00.29 go

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0 3834172   1,9g   5424 S  76,0   3,0   0:01.42 main
1660800 alex      20   0 1824120  16736   9928 S   0,0   0,0   0:00.29 go
...

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0   34,7g  20,4g   5400 S 192,0  32,6   0:38.15 main
1660800 alex      20   0 1824120  14240   8392 S   0,0   0,0   0:00.32 go
...

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0   34,7g  26,3g   5380 S  86,0  42,1   0:55.91 main
1660800 alex      20   0 1824120  11892   6052 S   0,0   0,0   0:00.33 go

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0   34,7g  26,7g   5380 S 100,0  42,8   0:56.91 main
1660800 alex      20   0 1824120  11892   6052 S   0,0   0,0   0:00.33 go

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0   34,7g  26,9g   5332 S   0,0  43,1   1:03.83 main
1660800 alex      20   0 1824120  11856   6016 S   0,0   0,0   0:00.34 go

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1660918 alex      20   0   34,7g  26,9g   5332 S   0,0  43,1   1:03.83 main
1660800 alex      20   0 1824120  11856   6016 S   0,0   0,0   0:00.34 go^C
```