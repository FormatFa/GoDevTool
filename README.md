Run:
go run .
打包:
- windows
 fyne package -os windows --icon res/logo.png

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
