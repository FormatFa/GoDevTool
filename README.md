Run:
go run .
 打包:
 - windows
  fyne package -os windows --icon res/logo.png 

# TUI 命令行版（端口占用查询/杀进程）
构建:
  go build -o gokit.exe ./cmd/gokit
运行:
  ./gokit.exe
按键:
  ↑/↓ 移动  d 杀进程(二次确认)  / 过滤端口  r 刷新  q 退出
说明:
  - 表格下方实时显示选中进程的完整命令行参数
  - 扫描/杀进程逻辑位于 internal/portscan，GUI 与 TUI 共用
  - 支持 Windows(netstat -ano)、macOS(lsof)、Linux(netstat -tlnp)
  - 杀系统进程失败时可能需要管理员/超级用户权限

# 可选实机测试
  $env:GO_KIT_LIVE_TEST='1'; go test ./internal/portscan

set ANDROID_HOME=E:\env\android\androidsdk
set ANDROID_NDK_HOME=E:\env\android\android-ndk-r25b-windows
fyne package -os android -appID com.example.myapp -icon res\logo.png
fyne package -os android

 压缩体积
 https://blog.csdn.net/qq_22598991/article/details/124046790


TODO:
- 菜单生成
- mvc?

- mac检测是否超级权限运行
