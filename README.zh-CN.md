# MusicUnlocker

[English](README.md) | **简体中文**

MusicUnlocker 是一款面向 Windows 11 x64 的桌面工具，可在本地解码受支持的加密音乐文件。项目采用共享的 Go 核心，以及基于 Wails v2、React 和 TypeScript 的桌面前端。图形界面（GUI）与开发者命令行工具（CLI）共用相同的文件探测、解码、批处理、命名及输出规则。

目前涉及的格式家族包括 KGM/VPR、NCM、KWM、QMC、TM，以及虾米音乐和喜马拉雅相关格式。**具备对应处理路径不等于所有真实文件变体都已通过验证**：部分变体仍待验证；QMC STag 或 KGM V5/KGG 输入需要在运行时明确提供兼容的外部材料。详情请参阅[真实文件验证矩阵](docs/real-validation.md)。

应用遵循**离线优先**原则：保留源文件、不在未告知的情况下覆盖输出文件、以原子方式写入解码结果，并避免不必要的音频转码。最终输出的文件扩展名根据解码后内容的实际格式识别，而不是根据加密文件原有的扩展名推断。元数据读取或文件命名失败时会安全回退，不会修改已解码的音频字节。

桌面应用支持原生文件/文件夹选择、递归批量处理、系统原生拖放、启动参数与单实例队列交接、根据元数据命名、界面本地化、安全的本地音频预览、外部材料补充后重试、跟随系统/浅色/深色主题，以及关于页面和版本信息展示。

## 开发者构建

安装仓库要求的 Go 和前端依赖，以及 Wails v2 命令行工具，然后执行：

```powershell
wails dev
wails build
```

架构与当前开发交接信息参见[GUI 架构](docs/gui-architecture.md)和[项目状态](docs/project-status.md)。计划事项统一记录在[路线图](docs/roadmap.md)中。

Windows 发布打包说明（包括 NSIS 前置依赖、便携版 ZIP、WebView2 处理策略及校验和工作流）参见 [Windows 发布文档](docs/windows-release.md)。

## 持续集成与候选发布版本

向维护中的开发分支提交 Pull Request 或推送代码时，会运行 Windows 持续集成（CI）工作流，包括 Go 测试与 `go vet`、锁定依赖的前端构建及冒烟测试，以及生产模式的 Wails 构建。

独立的 Windows Release Candidate（RC，候选发布版本）工作流需要手动触发。它在 `windows-2025` 环境中构建包含六个指定文件的正式发布目录，验证文件清单和 `SHA256SUMS`，并上传保留 30 天的工作流产物。

创建 GitHub Draft Release（草稿版本）是一个明确的手动操作，且必须事先存在对应标签：首个 RC 使用 `v0.7.0-rc.1`，后续使用 `-rc.N`；准备好正式版专属说明后，才可使用 `v0.7.0`。RC 草稿会明确标记为 GitHub Prerelease（预发布版本）。

产品版本、Windows PE 文件版本信息以及发布文件名继续使用 `0.7.0`。每个带标签的 RC 都必须具有对应的 `docs/releases/0.7.0-rc.N.md` 说明。工作流不会自动公开发布 Release、创建或移动 Git 标签，也不会给可执行文件进行代码签名。

工作流触发方式、权限、产物证明（Attestation）、代码签名及干净运行环境的验证边界，详见 [Windows 发布文档](docs/windows-release.md)。

## Windows 发行形式

`0.7.0` 候选发布工作流会生成三种 Windows x64 应用形式：

- **安装程序**：`MusicUnlocker-Setup-0.7.0-win64.exe`。支持当前用户安装、开始菜单入口、卸载，以及加密文件的文件关联。
- **独立 EXE**：`MusicUnlocker-0.7.0-win64.exe`。这是真正的单文件 Wails 应用，无需安装 MusicUnlocker 或解压即可直接运行。
- **便携版 ZIP**：`MusicUnlocker-0.7.0-win64-portable.zip`。包含可执行文件、README、许可证和第三方声明的完整压缩包。

独立 EXE 和便携版不会注册文件关联。三种形式使用相同的应用及版本标识，并采用官方内置 Evergreen WebView2 引导安装程序的恢复策略。

**功能范围说明：**MusicUnlocker 不是流媒体客户端、音乐下载器、云服务、强制转码工具，也不是完整的音乐媒体库管理器。
