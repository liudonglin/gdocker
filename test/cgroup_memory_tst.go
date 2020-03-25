package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

func LimitMemory() {
	if os.Args[0] == "/proc/self/exe" {

		//容器进程
		fmt.Printf("current 容器进程 pid %d", syscall.Getpid())
		fmt.Println()

		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	} else {
		// 得到fork出来进程映射在外部命名空间的pid
		fmt.Printf("外部命名空间的pid %v", cmd.Process.Pid)
		fmt.Println()

		//在系统默认创建挂载了 memory subsystem的Hierarchy上创建cgroup
		if _, err = os.Stat(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit")); err != nil {
			err = os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit"), 0755)
			if err != nil {
				fmt.Println("Mkdir testmemorylimit ERROR", err)
				os.Exit(1)
			}
		}

		//将容器进程加入到这个cgroup中
		_ = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
		//限制cgroup进程使用
		_ = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "memory.limit_in_bytes"), []byte("lOOm"), 0644)

		_, _ = cmd.Process.Wait()
	}

}
