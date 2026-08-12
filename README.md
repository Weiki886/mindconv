<div align="center">

# mindconv

**将 MindManager `.mmap` 文件转换为 Markdown、HTML 或原始 XML 的轻量级命令行工具。**

[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#安装)

[快速开始](#快速开始) · [使用说明](#使用说明) · [工作原理](#工作原理) · [参与贡献](#参与贡献)

</div>

`mindconv` 使用 Go 编写，在本地读取 MindManager 的 ZIP/XML 数据，不依赖 MindManager、浏览器或后台服务。它适合将思维导图整理为便于版本管理、发布和二次处理的文本格式。

## 功能亮点

- **多格式输出**：生成层级化 Markdown、响应式单文件 HTML，或原样导出 `Document.xml`。
- **内容提取**：保留中心主题、子主题、备注以及安全的 HTTP、HTTPS、邮件和相对链接。
- **本地处理**：文件不会上传到网络，转换过程完全在本机完成。
- **安全默认值**：限制 ZIP 条目数、XML 大小、节点数和嵌套深度，并过滤 `javascript:` 等危险链接。
- **脚本友好**：支持标准输出、自定义输出路径和明确的覆盖保护。
- **跨平台**：可在 macOS、Linux 和 Windows 上构建为单个可执行文件。

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [使用说明](#使用说明)
- [命令行参数](#命令行参数)
- [工作原理](#工作原理)
- [转换规则](#转换规则)
- [安全设计](#安全设计)
- [当前限制](#当前限制)
- [开发](#开发)
- [参与贡献](#参与贡献)
- [常见问题](#常见问题)
- [许可证](#许可证)

## 安装

### 使用 Go 安装

需要 Go 1.22 或更高版本：

```bash
go install github.com/Weiki886/mindconv/cmd/mindconv@latest
```

确保 Go 的二进制目录已经加入 `PATH`：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
mindconv --version
```

### 从源码构建

```bash
git clone https://github.com/Weiki886/mindconv.git
cd mindconv
mkdir -p bin
go build -o bin/mindconv ./cmd/mindconv
./bin/mindconv --help
```

构建完成后的二进制文件不依赖 Go、MindManager 或其他运行时。

## 快速开始

```bash
# 转换为 Markdown，生成 project.md
mindconv project.mmap

# 转换为 HTML，生成 project.html
mindconv --format html project.mmap

# 导出原始 Document.xml，生成 project.xml
mindconv --format xml project.mmap
```

默认输出文件与输入文件位于同一目录。程序不会覆盖已有文件，除非显式传入 `--force`。

## 使用说明

### 指定输出位置

```bash
mindconv project.mmap --output docs/project.md
mindconv project.mmap -f html -o public/project.html
mindconv project.mmap -f xml -o exports/project.xml
```

### 输出到标准输出

```bash
mindconv project.mmap --stdout
mindconv project.mmap -f html -o -
mindconv project.mmap -f xml --stdout
```

这适合与其他命令组合：

```bash
mindconv project.mmap --stdout | less
```

### 覆盖已有文件

```bash
mindconv project.mmap --force
```

## 命令行参数

```text
mindconv [options] <input.mmap>
```

| 参数 | 说明 |
| --- | --- |
| `-f, --format` | 输出格式：`md`、`markdown`、`html`、`htm` 或 `xml`；默认 `md` |
| `-o, --output` | 指定输出路径；使用 `-` 表示标准输出 |
| `--stdout` | 将结果写入标准输出 |
| `--force` | 覆盖已经存在的输出文件 |
| `--version` | 显示版本号 |
| `-h, --help` | 显示帮助信息 |

## 工作原理

`.mmap` 本质上是一个 ZIP 容器，其中的 `Document.xml` 保存思维导图核心数据。`mindconv` 不会把整个压缩包解压到磁盘，而是直接读取所需内容。

```text
                           ┌─→ Markdown
.mmap → Document.xml → Map ├─→ HTML
              │            └─→ 其他渲染器
              └──────────────→ 原始 XML
```

Markdown 和 HTML 先经过格式无关的内部模型，以隔离不同 MindManager 版本的 XML 差异；XML 模式则原样导出 `Document.xml`，不会重新格式化或序列化。

## 转换规则

| MMAP 内容 | Markdown | HTML | XML |
| --- | --- | --- | --- |
| 中心主题 | 一级标题 | 页面标题和根节点 | 保留原始节点 |
| 子主题 | 二至六级标题 | 嵌套主题卡片 | 保留原始节点 |
| 更深层主题 | 缩进列表 | 嵌套主题卡片 | 保留原始节点 |
| 备注 | 标题下方正文 | 主题备注区域 | 保留原始节点 |
| 超链接 | Markdown 链接 | 自动转义的 HTML 链接 | 保留原始节点 |

## 安全设计

`.mmap` 文件应被视为不可信输入。`mindconv` 当前采取以下措施：

- 不将 ZIP 内容直接解压到用户目录。
- 限制压缩包条目数以及 `Document.xml` 的解压大小。
- 限制 XML 节点数和最大嵌套深度。
- 不执行或自动打开附件。
- 只保留明确允许的链接协议。
- 使用 Go `html/template` 转义 HTML 动态内容。
- 默认拒绝覆盖已有输出文件。

XML 模式会保留源文件中的原始内容，其中可能包含备注、链接或其他敏感信息；共享导出的 XML 前请先检查并脱敏。

## 当前限制

- 仅支持 ZIP/XML 结构的 `.mmap`，不支持旧版 `.mmp`。
- 不还原自由布局、主题样式、边界框和关系线。
- 暂不导出嵌入图片和附件。
- Markdown 无法完整表达思维导图的图形布局，输出以内容层级为主。
- MindManager 不同版本可能存在 XML 差异；报告兼容问题时请提供脱敏后的最小样本和版本信息。

## 项目结构

```text
mindconv/
├── cmd/mindconv/       # CLI 入口
├── internal/
│   ├── cli/            # 参数解析、文件输出和退出码
│   ├── mmap/           # MMAP 容器与 XML 解析
│   ├── model/          # 格式无关的数据模型
│   └── render/         # Markdown 与 HTML 渲染
├── go.mod
├── LICENSE
└── README.md
```

## 开发

```bash
git clone https://github.com/Weiki886/mindconv.git
cd mindconv

gofmt -w cmd internal
go vet ./...
go test -race ./...
go build ./cmd/mindconv
```

解析器测试使用动态生成的最小 ZIP/XML 数据，不应提交包含个人或公司信息的真实 `.mmap` 文件。

## 参与贡献

欢迎提交 Issue 和 Pull Request：

1. 从最新的 `main` 创建功能分支。
2. 保持改动范围清晰，并为行为变化补充测试。
3. 确保格式检查、静态检查、测试和构建全部通过。
4. 使用清晰的 Conventional Commits 风格提交信息。

- [报告问题](https://github.com/Weiki886/mindconv/issues/new)
- [查看 Pull Requests](https://github.com/Weiki886/mindconv/pulls)

## 常见问题

<details>
<summary><strong>终端提示 <code>command not found: mindconv</code></strong></summary>

确认程序已经安装，并检查 Go 的二进制目录：

```bash
go install github.com/Weiki886/mindconv/cmd/mindconv@latest
export PATH="$(go env GOPATH)/bin:$PATH"
command -v mindconv
```

macOS 使用 zsh 时，可将下面这行加入 `~/.zshrc`，然后重新打开终端：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

</details>

<details>
<summary><strong><code>go run example.mmap</code> 提示 <code>outside main module</code></strong></summary>

`go run` 接收的是 Go 程序路径，而不是 `.mmap` 文件。请在源码目录中运行：

```bash
go run ./cmd/mindconv example.mmap
```

</details>

<details>
<summary><strong>如何卸载？</strong></summary>

先找到正在使用的程序：

```bash
command -v mindconv
```

如果通过 `go install` 安装，可删除对应的二进制文件：

```bash
mindconv_bin_dir="$(go env GOBIN)"
if [ -z "$mindconv_bin_dir" ]; then
  mindconv_bin_dir="$(go env GOPATH)/bin"
fi
rm "$mindconv_bin_dir/mindconv"
hash -r
```

删除程序通常不需要移除 `~/.zshrc` 中的 Go `PATH` 配置，因为其他 Go 工具也可能依赖该目录。

</details>

## 许可证

本项目基于 [MIT License](LICENSE) 发布。
