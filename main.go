package main

import (
	"fmt"
	"os"
	"path/filepath"
	"ppa/gui"
	_ "ppa/job"
	"strings"

	"github.com/btagrass/gobiz/cmd"
	"github.com/btagrass/gobiz/utl"
	"github.com/spf13/cobra"
)

// 入口
func main() {
	cmd.Execute(
		"ppa",
		"PhotoPrism Assistant",
		&cobra.Command{
			Use:   "install",
			Short: "安装",
			Run: func(c *cobra.Command, args []string) {
				// 安装软件
				_, err := utl.Command("sudo apt install ffmpeg -y")
				if err != nil {
					fmt.Println(err)
					return
				}
				// 添加配置
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Println(err)
					return
				}
				desktop := "Desktop"
				if strings.Contains(os.Getenv("LANG"), "zh") {
					desktop = "桌面"
				}
				dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
				if err != nil {
					fmt.Println(err)
					return
				}
				err = os.WriteFile(
					filepath.Join(home, desktop, "ppa.desktop"),
					[]byte(fmt.Sprintf(`
[Desktop Entry]
Name=PhotoPrism Assistant
Type=Application
Exec=%s/suc
Icon=%s/ico.png
Version=1.0.0
`, dir, dir)),
					os.ModePerm,
				)
				if err != nil {
					fmt.Println(err)
					return
				}
			},
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "卸载",
			Run: func(c *cobra.Command, args []string) {
				// 删除配置
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Println(err)
					return
				}
				desktop := "Desktop"
				if strings.Contains(os.Getenv("LANG"), "zh") {
					desktop = "桌面"
				}
				err = utl.Remove(filepath.Join(home, desktop, "ppa.desktop"))
				if err != nil {
					fmt.Println(err)
					return
				}
			},
		},
		&cobra.Command{
			Use:   "run",
			Short: "运行",
			Run: func(c *cobra.Command, args []string) {
				// 界面服务
				win := gui.NewMainGui()
				win.ShowAndRun()
			},
		},
	)
}
