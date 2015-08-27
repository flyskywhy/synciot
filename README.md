# App 编译方法

## 注意事项
由于 Android Studio 自身的 Bug ，如果在操作串口时发生 APP 闪退，且在 logcat 中看到`java.lang.UnsatisfiedLinkError: dlopen failed: cannot locate symbol "tcgetattr"`这样的出错信息，则说明碰到了[http://stackoverflow.com/questions/28740315/android-ndk-getting-java-lang-unsatisfiedlinkerror-dlopen-failed-cannot-loca](http://stackoverflow.com/questions/28740315/android-ndk-getting-java-lang-unsatisfiedlinkerror-dlopen-failed-cannot-loca)这样的 Bug 。在该 Bug 未修复前，可用如下方法绕过去：

在最后生成 apk 准备调试前，修改 `app/build.gradle` 中的`compileSdkVersion = 21`为`compileSdkVersion = 15`。后续如果修改代码会编译出错的，再修改回来。如此反复即可。
