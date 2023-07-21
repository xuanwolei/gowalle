# Go-Walle

Go-Walle（瓦力）：Android Signature V2 Scheme签名下的渠道/自定义数据写入

此项目为[Walle](https://github.com/Meituan-Dianping/walle)的golang版本核心算法实现，可对apk写入/读取自定义数据，与Walle Java版本算法兼容

通过在Apk中的`APK Signature Block`区块添加自定义的数据（渠道、自定义信息等），可应用于多渠道打包、apk自定义数据场景。

## Quick Start
#### 获取包
```
go get -u github.com/xuanwolei/gowalle
```

#### 写入数据
```
filePath := "./test.apk"
err := gowalle.WriteBlockByte(filePath, []byte("this is custom information 1 2 3 4 5!"))
```

#### 读取数据
```
filePath := "./test.apk"
data, err := gowalle.GetBlockByte(filePath)
if err != nil {
    return
}
fmt.Printf("block string:%s", string(data))
```
> 数据写入格式和Walle兼容，walle java代码也可读取

## 参考
* [Walle](https://github.com/Meituan-Dianping/walle)
* [MCRelease](https://github.com/LeoExer/MCRelease)
* [Android Source Code: apksig](https://android.googlesource.com/platform/tools/apksig/)
* [APK Signature Scheme v2](https://source.android.com/security/apksigning/v2.html)
* [Zip Format](https://en.wikipedia.org/wiki/Zip_(file_format))
* [Android Source Code: ApkSigner](https://android.googlesource.com/platform/build/+/8740e9d)
* [Android Source Code: apksig](https://android.googlesource.com/platform/tools/apksig/)