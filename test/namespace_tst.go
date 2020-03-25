package test

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// uts namespace主要用来隔离nodename和domainname,每个uts namespace允许拥有自己的hostname
func CreateNEWUTS() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	//pstree -pl 查看系统进程之间关系
	//---go(xxxx)-+-main(19912)
	//输出当前pid
	// echo $$
	// 19915
	// 验证一下父进程和子进程是否不在同一个UTS Namespace中，验证代码如下。
	//# readlink /proc/19912/ns/uts
	//uts: [4026531838]
	//# readlink /proc/19915/ns/uts
	//uts: [4026532193]

	//#修改hostname为bird然后打印出来.
	//# hostname -b bird
	//# hostname
	//另外启动一个shell，在宿主机上运行hostname，看一下效果
	//root@iZ254rt8xf1Z:~# hostname
	//iZ254rt8xf1Z
}

// IPC Namespace用来隔离System V IPC和POSIX message queues
//每一个IPC Namespace 都有自己的System V IPC和POSIX message queue
func CreateNEWIPC() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// PID Namespace是用来隔离进程ID的。同样一个进程在不同的PID Namespace里可以拥有不同的PID
// 这样就可以理解，在docker container里面，使用ps -ef经常会发现，在容器 内，前台运行的那个进程PID是1,但是在容器外，使用ps -ef会发现同样的进程却有不同的 PID
func CreateNEWPID() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	// 我们需要打开两个shell。首先在宿主机上看一下进程树，找一下进程的真实PID。
	// pstree -pl
	// 可以看到，go main函数运行的PID为20190
	// 下面，打开另外一个shell运行一下如下代码。
	// go run main.go
	// echo $$
	// 1
	// 可以看到，该操作打印了当前Namespace的PID，其值为1。

}

// Mount Namespace用来隔离各个进程看到的挂载点视图。在不同Namespace的进程中，看 到的文件系统层次是不一样的
// 在Mount Namespace中调用mount()和umount()仅仅只会影响 当前Namespace内的文件系统，而对全局的文件系统是没有影响的。
// chroot也是将某一个子目录变成根节点。但是，Mount Namespace不仅能实现这个功能，而且能以更加灵活和安全的方式实现。
// Mount Namespace是Linux第一个实现的Namespace类型，因此，它的系统调用参数 是NEWNS (New Namespace的缩写)。当时人们貌似没有意识到，以后还会有很多类型的 Namespace加入Linux大家庭
func CreateNEWNS() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	// 运行代码，然后查看一下/proc的文件内容
	// proc是一个文件系统，提供额外的机制， 可以通过内核和内核模块将信息发送给进程
	// ls /proc

}

// User Namespace主要是隔离用户的用户组ID。
// 一个进程的User ID和Group ID在User Namespace内外可以是不同的
// 比较常用的是，在宿主机上以一个非root用户运行 创建一个User Namespace,然后在User Namespace里面却映射成root用户。这意味着，这个 进程在User Namespace里面有root 权限，
// 但是在User Namespace外面却没有root 的权限
// 从LinuxKernel3.8开始，非root进程也可以创建UserNamespace，并且此用户在Namespace里 面可以被映射成root，且在Namespace内有root权限。
func CreateNEWUSER() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
	}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(1), Gid: uint32(1)}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	//运行前在宿主机上看一下当前的用户和用户组，显示如下。
	//root@ iz254rt8xf1Z:~/gocode/src/book# id
	//uid=0 (root) gid=0(root) groups=0 (root)
	//可以看到我们是root用户，接下来运行一下程序。
	//root@iZ254rt8xf1Z: ~/gocode/src/book# go run main.go
	//$ id
	//uid=65534 (nobody) gid= 65534 (nogroup) groups= 65534 (nogroup)
	//可以看到，它们的UID是不同的，因此说明User Namespace生效了
}

// Network Namespace是用来隔离网络设备、IP地址端口等网络栈的Namespace。Network
// Namespace可以让每个容器拥有自己独立的(虚拟的)网络设备，而且容器内的应用可以绑定
// 到自己的端口，每个Namespace内的端口都不会互相冲突。在宿主机上搭建网桥后，就能很方
// 便地实现容器之间的通信，而且不同容器上的应用可以使用相同的端口。
func CreateNEWNET() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(1), Gid: uint32(1)}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
