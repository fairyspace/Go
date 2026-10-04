#!/usr/bin/env python3
"""Build a self-contained, offline-friendly HTML companion for 教学文稿.

No third-party packages are required. Markdown remains the source of truth;
chapter HTML is generated from it, with the runnable Go examples embedded.
"""
from __future__ import annotations

import hashlib
import html
import os
import re
from pathlib import Path
from urllib.parse import quote, unquote, urlsplit, urlunsplit

SOURCE_DIR = Path(__file__).resolve().parent
REPO_ROOT = SOURCE_DIR.parents[1]
OUTPUT_DIR = SOURCE_DIR / "html"
CODE_ROOT = REPO_ROOT / "00-基础语法" / "codes" / "教学"

LESSONS = [
    {
        "id": "00", "file": "00-索引与学习路线.md", "page": "00-roadmap.html",
        "name": "索引与学习路线", "phase": "探险地图", "tagline": "把 21 篇散记，走成一条从能写到写快的 Go 学习路线。",
        "memory": "先能跑，再能写；写对、写好，最后写快。",
        "prompt": "不看目录，你能按顺序说出学习的四个阶段吗？",
        "hint": "能写 → 写对 → 写好 → 写快。先建立地图，再挑一篇动手。",
        "code_groups": [],
    },
    {
        "id": "01", "file": "01-环境安装与项目结构.md", "page": "01-setup.html",
        "name": "环境安装与项目结构", "phase": "开机起步", "tagline": "从安装 Go 到跑通 Hello World，再把项目摆放得清清楚楚。",
        "memory": "装好 Go → `go version` 验身份 → `go mod init` 开新项目 → `go run` 点火。",
        "prompt": "新建一个空文件夹，能不能在 30 秒内跑出 Hello World？",
        "hint": "GOROOT 不折腾；项目从 go.mod 开始。先运行 `go version`，再创建模块。",
        "code_groups": [
            {"title": "第一声 Hello", "run": "go run ./s01_hello", "files": ["s01_hello/main.go"]},
        ],
    },
    {
        "id": "02", "file": "02-变量、常量与数据类型.md", "page": "02-values.html",
        "name": "变量、常量与数据类型", "phase": "开机起步", "tagline": "给数据贴上类型标签，学会声明、交换与格式化输出。",
        "memory": "`const` 固定不动，`var` 有零值；`:=` 是函数里的快捷新朋友。",
        "prompt": "变量不赋值会怎样？常量可以只写类型、不写初始值吗？",
        "hint": "零值属于变量；常量必须有初始表达式。`:=` 只能在函数体中使用。",
        "code_groups": [
            {"title": "变量、零值、iota 与 fmt", "run": "go run ./s02_var", "files": ["s02_var/main.go"]},
        ],
    },
    {
        "id": "03", "file": "03-数组、切片与 Map.md", "page": "03-collections.html",
        "name": "数组、切片与 Map", "phase": "给数据安家", "tagline": "三种容器，各有脾气：固定格子、可伸缩窗口、键值字典。",
        "memory": "数组是固定格子；切片是数组上的窗口；Map 是查名字的字典。`len` 看眼前，`cap` 看后备。",
        "prompt": "对切片 append 后，为什么通常要把结果重新赋回原变量？",
        "hint": "切片像窗口；扩容可能换一张底层数组，返回的新窗口要接住。",
        "code_groups": [
            {"title": "数组：固定长度的格子", "run": "go run ./s03_array", "files": ["s03_array/main.go"]},
            {"title": "切片：会长大的窗口", "run": "go run ./s04_slice", "files": ["s04_slice/main.go"]},
            {"title": "Map：键值字典", "run": "go run ./s06_map", "files": ["s06_map/main.go"]},
        ],
    },
    {
        "id": "04", "file": "04-结构体与 JSON.md", "page": "04-struct-json.html",
        "name": "结构体与 JSON", "phase": "给数据安家", "tagline": "用 struct 画好数据表单，再用 JSON 和外部世界握手。",
        "memory": "`struct` 是表单蓝图，JSON tag 是双语贴纸；ID 要精确，别随手塞进 `float64`。",
        "prompt": "JSON 里的字段名和 Go 字段名不同时，靠什么对上？",
        "hint": "给导出的 struct 字段加上 `json:\"name\"` 这样的 tag。",
        "code_groups": [
            {"title": "结构体、方法与 JSON tag", "run": "go run ./s05_struct", "files": ["s05_struct/main.go"]},
            {"title": "强类型 / 弱类型 JSON 解析", "run": "go run ./s12_json", "files": ["s12_json/main.go"]},
        ],
    },
    {
        "id": "05", "file": "05-循环与流程控制.md", "page": "05-control-flow.html",
        "name": "循环与流程控制", "phase": "让代码动起来", "tagline": "让程序按规则前进、跳过、分支，写出清楚的执行路径。",
        "memory": "Go 的循环一把 `for` 梭；`continue` 跳过这轮，`break` 离开当前循环。",
        "prompt": "`switch` 默认会自动落到下一个 case 吗？`fallthrough` 呢？",
        "hint": "默认不会；只有显式写 `fallthrough` 才会继续执行下一个 case。",
        "code_groups": [
            {"title": "for、range、switch 与跳转", "run": "go run ./s07_loop", "files": ["s07_loop/main.go"]},
        ],
    },
    {
        "id": "06", "file": "06-函数、闭包与 defer.md", "page": "06-functions.html",
        "name": "函数、闭包与 defer", "phase": "让代码动起来", "tagline": "把逻辑打包复用，理解闭包捕获，以及 defer 的执行时机。",
        "memory": "函数先入场，`defer` 后入先出；defer 的参数先记下，函数返回时再执行。",
        "prompt": "同一个函数里注册了三个 defer，执行顺序是怎样的？",
        "hint": "像叠盘子：后放上去的先拿走，也就是 LIFO。",
        "code_groups": [
            {"title": "函数、闭包与错误处理", "run": "go run ./s08_func", "files": ["s08_func/main.go"]},
            {"title": "defer 经典题与资源释放", "run": "go run ./s08_defer", "files": ["s08_defer/main.go"]},
            {"title": "签名函数小实战", "run": "go run ./s08_sign", "files": ["s08_sign/main.go"]},
        ],
    },
    {
        "id": "07", "file": "07-接口与 Options 模式.md", "page": "07-interfaces.html",
        "name": "接口与 Options 模式", "phase": "让代码动起来", "tagline": "让代码依赖“能做什么”，而不是“具体是谁”，再设计好可选配置。",
        "memory": "接口看能力，不看血统：方法集齐了，就能上场。",
        "prompt": "一个类型要显式写 `implements` 才能实现 Go 接口吗？",
        "hint": "不需要。Go 接口是隐式实现：类型拥有接口要求的方法，就满足它。",
        "code_groups": [
            {"title": "隐式实现与类型断言", "run": "go run ./s09_interface", "files": ["s09_interface/main.go", "s09_interface/study/study.go"]},
            {"title": "Options 模式与配置选项", "run": "go run ./s10_option", "files": ["s10_option/main.go", "s10_option/friend/friend.go", "s10_option/friend/option.go"]},
        ],
    },
    {
        "id": "08", "file": "08-并发编程.md", "page": "08-concurrency.html",
        "name": "并发编程", "phase": "并发冲刺", "tagline": "让多个任务安全协作：用 channel 传话、WaitGroup 等待、锁保护共享数据。",
        "memory": "goroutine 负责跑，channel 负责传话，WaitGroup 负责等收工；共享 Map 要有保护。",
        "prompt": "`wg.Add(1)` 应该在启动 goroutine 前还是后？谁负责 close channel？",
        "hint": "先 Add 再 go；发送方（通常是生产者）负责 close，接收方不要抢着关。",
        "code_groups": [
            {"title": "Channel、select 与 Worker Pool", "run": "go run ./s13_chan", "files": ["s13_chan/main.go"]},
            {"title": "WaitGroup 与任务编排", "run": "go run ./s14_waitgroup", "files": ["s14_waitgroup/main.go", "s14_waitgroup/context.go"]},
            {"title": "sync.Map 与安全 Map", "run": "go run ./s15_syncmap", "files": ["s15_syncmap/main.go"]},
        ],
    },
    {
        "id": "09", "file": "09-性能调优.md", "page": "09-performance.html",
        "name": "性能调优", "phase": "并发冲刺", "tagline": "先测量，再优化：用基准测试找到热点，再看逃逸和内存分配。",
        "memory": "先拿证据（Benchmark），再动手；`sync.Pool` 是临时周转箱，不是长期仓库。",
        "prompt": "优化代码前，先用什么证明它真的慢？改完之后又看哪些指标？",
        "hint": "先跑 Benchmark 建立基线；对比 ns/op、B/op、allocs/op，并在目标机器复测。",
        "code_groups": [
            {"title": "Benchmark 与加密 / 字符串对比", "run": "go test -bench . -benchmem ./s16_bench", "files": ["s16_bench/crypto.go", "s16_bench/crypto_test.go"]},
            {"title": "逃逸分析小实验", "run": "go build -gcflags '-m -l' ./s17_escape", "files": ["s17_escape/main.go"]},
            {"title": "sync.Pool 复用与基准", "run": "go run ./s17_pool  ·  go test -bench . -benchmem ./s17_pool", "files": ["s17_pool/main.go", "s17_pool/pool_test.go"]},
        ],
    },
    {
        "id": "10", "file": "10-常用标准库速查.md", "page": "10-stdlib.html",
        "name": "常用标准库速查", "phase": "并发冲刺", "tagline": "把时间、字符串、转换和错误处理这些高频工具收入自己的工具箱。",
        "memory": "Go 时间布局不是模板，是“参考时刻”：2006-01-02 15:04:05。",
        "prompt": "把 `2020-11-08T08:18:46+08:00` 转成常见日期格式，第一步是什么？",
        "hint": "先按 `time.RFC3339` Parse 成 time.Time，再用目标 layout Format。",
        "code_groups": [
            {"title": "time.RFC3339、时区与常用时间操作", "run": "go run ./s11_timeutil", "files": ["s11_timeutil/main.go", "s11_timeutil/timeutil/timeutil.go"]},
        ],
    },
]

PHASES = [
    ("01", "开机起步", "先把工具装好，再学会让数据有名字。", ["01", "02"], "🧭"),
    ("02", "给数据安家", "数组、切片、Map 与结构体，数据住进合适的容器。", ["03", "04"], "📦"),
    ("03", "让代码动起来", "流程控制、函数、接口——让程序按你的设计行动。", ["05", "06", "07"], "✏️"),
    ("04", "并发冲刺", "并发协作、性能测量与标准库，写得稳也写得快。", ["08", "09", "10"], "🚀"),
]

LESSON_BY_ID = {lesson["id"]: lesson for lesson in LESSONS}
SOURCE_TO_PAGE = {
    (SOURCE_DIR / lesson["file"]).resolve(): lesson["page"] for lesson in LESSONS
}


def esc(value: object) -> str:
    return html.escape(str(value), quote=True)


def relative_url(url: str, source: Path) -> str:
    """Rebase relative Markdown links from the source folder to html/."""
    raw = url.strip()
    if not raw or raw.startswith(("#", "//", "http://", "https://", "mailto:", "tel:", "data:")):
        return raw
    parts = urlsplit(raw)
    if not parts.path:
        return raw
    decoded_path = unquote(parts.path)
    if decoded_path.startswith("/"):
        target = (REPO_ROOT / decoded_path.lstrip("/")).resolve()
    else:
        target = (source.parent / decoded_path).resolve()
    page = SOURCE_TO_PAGE.get(target)
    if page:
        rebased = os.path.relpath(OUTPUT_DIR / page, OUTPUT_DIR).replace(os.sep, "/")
    else:
        rebased = os.path.relpath(target, OUTPUT_DIR).replace(os.sep, "/")
        if parts.path.endswith("/") and not rebased.endswith("/"):
            rebased += "/"
    return urlunsplit(("", "", quote(rebased, safe="/-._~%"), parts.query, parts.fragment))


def inline_markdown(text: str, source: Path) -> str:
    """Render the inline Markdown used in these notes without external tools."""
    saved: list[str] = []

    def stash(fragment: str) -> str:
        token = f"@@HTMLTOKEN{len(saved)}@@"
        saved.append(fragment)
        return token

    # Protect code spans first so that markup inside them stays literal.
    text = re.sub(r"(`+)(.+?)\1", lambda m: stash(f"<code>{esc(m.group(2))}</code>"), text)

    link_pattern = re.compile(r"(!?)\[([^\]]*)\]\(([^)\s]+)(?:\s+\"([^\"]*)\")?\)")

    def link_replace(match: re.Match[str]) -> str:
        is_image = match.group(1) == "!"
        label = match.group(2)
        destination = relative_url(match.group(3), source)
        title = match.group(4)
        if is_image:
            fragment = f'<img class="inline-image" src="{esc(destination)}" alt="{esc(label)}" loading="lazy"'
            if title:
                fragment += f' title="{esc(title)}"'
            fragment += ">"
        else:
            external = destination.startswith(("http://", "https://", "mailto:"))
            fragment = f'<a href="{esc(destination)}"'
            if external:
                fragment += ' target="_blank" rel="noopener noreferrer"'
            if title:
                fragment += f' title="{esc(title)}"'
            fragment += f">{html.escape(label, quote=False)}</a>"
        return stash(fragment)

    text = link_pattern.sub(link_replace, text)
    safe = html.escape(text, quote=False)
    safe = re.sub(r"\*\*(.+?)\*\*", r"<strong>\1</strong>", safe)
    safe = re.sub(r"__(.+?)__", r"<strong>\1</strong>", safe)
    safe = re.sub(r"~~(.+?)~~", r"<del>\1</del>", safe)
    safe = re.sub(r"(?<!\*)\*([^*\n]+)\*(?!\*)", r"<em>\1</em>", safe)
    safe = re.sub(r"(?<!_)_([^_\n]+)_(?!_)", r"<em>\1</em>", safe)
    safe = re.sub(r"(?<![\w\"=])((?:https?://)[^\s<]+)", r'<a href="\1" target="_blank" rel="noopener noreferrer">\1</a>', safe)
    for index, fragment in enumerate(saved):
        safe = safe.replace(f"@@HTMLTOKEN{index}@@", fragment)
    return safe


def plain_text(rendered: str) -> str:
    return re.sub(r"\s+", " ", re.sub(r"<[^>]+>", "", rendered)).strip()


def is_table_start(lines: list[str], index: int) -> bool:
    if index + 1 >= len(lines) or "|" not in lines[index]:
        return False
    return bool(re.match(r"^\s*\|?\s*:?-{3,}:?\s*(?:\|\s*:?-{3,}:?\s*)+\|?\s*$", lines[index + 1]))


def split_table_row(line: str) -> list[str]:
    line = line.strip()
    if line.startswith("|"):
        line = line[1:]
    if line.endswith("|"):
        line = line[:-1]
    return [cell.strip() for cell in line.split("|")]


def render_mermaid(source_code: str) -> str:
    """Render the small Mermaid flowchart subset used by the manuscripts as SVG."""
    direction_match = re.search(r"\bgraph\s+(LR|RL|TD|TB|BT)\b", source_code)
    direction = direction_match.group(1) if direction_match else "LR"
    horizontal = direction in ("LR", "RL")
    nodes: dict[str, dict[str, str]] = {}
    edges: list[tuple[str, str, str]] = []
    node_re = re.compile(r"([A-Za-z][A-Za-z0-9_]*)\s*(\[[^\]]*\]|\{[^}]*\})")

    for raw_line in source_code.splitlines():
        line = raw_line.strip()
        if not line or line.startswith(("graph ", "style ", "classDef ", "%%")):
            continue
        for match in node_re.finditer(line):
            node_id = match.group(1)
            shape = "decision" if match.group(2).startswith("{") else "normal"
            label = match.group(2)[1:-1].strip().strip('"').strip("'")
            label = label.replace("<br/>", "\n").replace("<br>", "\n").replace("&amp;", "&")
            nodes.setdefault(node_id, {"label": label, "shape": shape})
        if "-->" in line:
            parts = re.split(r"\s*-->\s*", line)
            parsed: list[tuple[str, str]] = []
            for part in parts:
                edge_label = ""
                label_match = re.match(r"^\|([^|]*)\|\s*(.*)$", part.strip())
                if label_match:
                    edge_label = label_match.group(1).strip().strip('"').strip("'").replace("<br/>", " · ").replace("<br>", " · ")
                    part = label_match.group(2)
                id_match = re.match(r"\s*([A-Za-z][A-Za-z0-9_]*)", part)
                if id_match:
                    parsed.append((id_match.group(1), edge_label))
            for left, right in zip(parsed, parsed[1:]):
                edges.append((left[0], right[0], right[1]))
                nodes.setdefault(left[0], {"label": left[0], "shape": "normal"})
                nodes.setdefault(right[0], {"label": right[0], "shape": "normal"})

    if not nodes:
        return f'<div class="code-card"><pre><code class="language-mermaid">{esc(source_code)}</code></pre></div>'

    # Assign columns/rows using a topological longest-path pass.
    incoming = {key: 0 for key in nodes}
    outgoing: dict[str, list[str]] = {key: [] for key in nodes}
    for start, end, _ in edges:
        incoming[end] = incoming.get(end, 0) + 1
        outgoing.setdefault(start, []).append(end)
    order_index = {key: index for index, key in enumerate(nodes)}
    queue = [key for key in nodes if incoming.get(key, 0) == 0]
    layers = {key: 0 for key in nodes}
    visited: list[str] = []
    while queue:
        current = queue.pop(0)
        visited.append(current)
        for child in outgoing.get(current, []):
            layers[child] = max(layers.get(child, 0), layers[current] + 1)
            incoming[child] -= 1
            if incoming[child] == 0:
                queue.append(child)
    for key in nodes:
        if key not in visited:
            layers[key] = max(layers.values(), default=0) + order_index[key] / max(len(nodes), 1)

    layer_groups: dict[float, list[str]] = {}
    for key in nodes:
        layer_groups.setdefault(layers[key], []).append(key)
    sorted_layers = sorted(layer_groups)
    for group in layer_groups.values():
        group.sort(key=lambda key: order_index[key])

    node_w, node_h = 184, 76
    gap_x, gap_y, margin = 82, 28, 28
    max_group_size = max(len(group) for group in layer_groups.values())
    if horizontal:
        width = 2 * margin + len(sorted_layers) * node_w + max(0, len(sorted_layers) - 1) * gap_x
        height = 2 * margin + max_group_size * node_h + max(0, max_group_size - 1) * gap_y
    else:
        width = 2 * margin + max_group_size * node_w + max(0, max_group_size - 1) * gap_x
        height = 2 * margin + len(sorted_layers) * node_h + max(0, len(sorted_layers) - 1) * gap_y
    positions: dict[str, tuple[float, float]] = {}
    for layer_index, layer in enumerate(sorted_layers):
        group = layer_groups[layer]
        for position_index, node_id in enumerate(group):
            if horizontal:
                x = margin + layer_index * (node_w + gap_x)
                group_h = len(group) * node_h + max(0, len(group) - 1) * gap_y
                y = margin + (height - 2 * margin - group_h) / 2 + position_index * (node_h + gap_y)
            else:
                y = margin + layer_index * (node_h + gap_y)
                group_w = len(group) * node_w + max(0, len(group) - 1) * gap_x
                x = margin + (width - 2 * margin - group_w) / 2 + position_index * (node_w + gap_x)
            positions[node_id] = (x, y)

    marker_id = "arrowhead-" + hashlib.sha1(source_code.encode("utf-8")).hexdigest()[:8]
    svg = [
        f'<svg class="mermaid-svg" width="{width:.0f}" height="{height:.0f}" viewBox="0 0 {width:.0f} {height:.0f}" role="img" aria-label="流程图（手绘风格）" xmlns="http://www.w3.org/2000/svg">',
        f'<defs><marker id="{marker_id}" markerWidth="9" markerHeight="9" refX="7" refY="4.5" orient="auto"><path d="M1 1 L8 4.5 L1 8" fill="none" stroke="#52655f" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></marker></defs>',
    ]
    for edge_index, (start, end, label) in enumerate(edges):
        if start not in positions or end not in positions:
            continue
        sx, sy = positions[start]
        tx, ty = positions[end]
        if horizontal:
            x1, y1, x2, y2 = sx + node_w, sy + node_h / 2, tx, ty + node_h / 2
            ctrl = max(24, (x2 - x1) * 0.45)
            path = f"M{x1:.1f},{y1:.1f} C{x1 + ctrl:.1f},{y1:.1f} {x2 - ctrl:.1f},{y2:.1f} {x2:.1f},{y2:.1f}"
            lx, ly = (x1 + x2) / 2, (y1 + y2) / 2 - 7
        else:
            x1, y1, x2, y2 = sx + node_w / 2, sy + node_h, tx + node_w / 2, ty
            ctrl = max(24, (y2 - y1) * 0.45)
            path = f"M{x1:.1f},{y1:.1f} C{x1:.1f},{y1 + ctrl:.1f} {x2:.1f},{y2 - ctrl:.1f} {x2:.1f},{y2:.1f}"
            lx, ly = (x1 + x2) / 2 + 10, (y1 + y2) / 2
        svg.append(f'<path class="diagram-edge" d="{path}" marker-end="url(#{marker_id})"/>')
        if label:
            svg.append(f'<text class="diagram-edge-label" x="{lx:.1f}" y="{ly:.1f}" text-anchor="middle">{esc(label)}</text>')

    palette = ["#e5f1eb", "#fff0c2", "#ffe3d6", "#e6edf8", "#eee7f4", "#e4f2f4"]
    for node_index, (node_id, item) in enumerate(nodes.items()):
        x, y = positions[node_id]
        color = palette[node_index % len(palette)]
        label_lines = item["label"].splitlines() or [node_id]
        font_size = 14 if len(label_lines) <= 2 else 12
        line_height = 18
        start_y = y + node_h / 2 - ((len(label_lines) - 1) * line_height) / 2 + 4
        if item["shape"] == "decision":
            points = f"{x + node_w / 2},{y + 2} {x + node_w - 2},{y + node_h / 2} {x + node_w / 2},{y + node_h - 2} {x + 2},{y + node_h / 2}"
            svg.append(f'<polygon class="diagram-node decision-node" points="{points}" fill="{color}"/>')
        else:
            svg.append(f'<rect class="diagram-node" x="{x + 2:.1f}" y="{y + 3:.1f}" width="{node_w - 3}" height="{node_h - 4}" rx="17" fill="{color}"/>')
            svg.append(f'<path class="diagram-node-echo" d="M{x + 14:.1f},{y + 5:.1f} Q{x + node_w / 2:.1f},{y - 1:.1f} {x + node_w - 10:.1f},{y + 5:.1f}"/>')
        svg.append(f'<text class="diagram-node-text" x="{x + node_w / 2:.1f}" y="{start_y:.1f}" text-anchor="middle" font-size="{font_size}">')
        for line_index, line in enumerate(label_lines):
            svg.append(f'<tspan x="{x + node_w / 2:.1f}" dy="{0 if line_index == 0 else line_height}">{esc(line)}</tspan>')
        svg.append("</text>")
    svg.append("</svg>")
    return '<div class="diagram-scroll"><div class="diagram-paper">' + "".join(svg) + '</div></div>'


def render_code_block(code: str, language: str = "") -> str:
    lang = language.strip().lower() or "text"
    label = lang.upper() if lang != "text" else "代码片段"
    return (
        '<div class="code-card">'
        f'<div class="code-toolbar"><span class="code-language">{esc(label)}</span>'
        '<button class="copy-code" type="button" aria-label="复制代码">复制</button></div>'
        f'<pre><code class="language-{esc(lang)}">{esc(code)}</code></pre>'
        '</div>'
    )


def list_marker(line: str):
    return re.match(r"^(\s*)([-+*]|\d+[.)])\s+(.*)$", line)


def render_markdown(markdown: str, source: Path, heading_items: list[dict] | None = None) -> str:
    """Render the repository's Markdown subset: fences, headings, tables, lists, quotes."""
    headings = heading_items if heading_items is not None else []
    lines = markdown.replace("\r\n", "\n").replace("\r", "\n").split("\n")
    output: list[str] = []
    i = 0

    def starts_block(index: int) -> bool:
        line = lines[index]
        return bool(
            re.match(r"^\s*(```|~~~)", line)
            or re.match(r"^\s{0,3}#{1,6}\s+", line)
            or re.match(r"^\s{0,3}>", line)
            or re.match(r"^\s{0,3}(?:[-*_]\s*){3,}$", line)
            or list_marker(line)
            or is_table_start(lines, index)
        )

    while i < len(lines):
        line = lines[i]
        if not line.strip():
            i += 1
            continue

        fence = re.match(r"^\s*(```+|~~~+)([^`]*)$", line)
        if fence:
            marker, language = fence.group(1), fence.group(2).strip()
            code_lines: list[str] = []
            i += 1
            while i < len(lines) and not re.match(rf"^\s*{re.escape(marker[0])}{{{len(marker)},}}\s*$", lines[i]):
                code_lines.append(lines[i])
                i += 1
            if i < len(lines):
                i += 1
            code = "\n".join(code_lines)
            if language.lower() == "mermaid":
                output.append('<div class="diagram-card"><div class="diagram-label">✎ 路线草图</div>')
                output.append(render_mermaid(code))
                output.append('<details class="diagram-source"><summary>看看这张图的 Mermaid 源码</summary>')
                output.append(render_code_block(code, "mermaid"))
                output.append('</details></div>')
            else:
                output.append(render_code_block(code, language))
            continue

        if re.match(r"^\s{0,3}#{1,6}\s+", line):
            match = re.match(r"^\s{0,3}(#{1,6})\s+(.*?)\s*#*\s*$", line)
            level = len(match.group(1))
            title = match.group(2).strip()
            heading_id = f"section-{len(headings) + 1:02d}"
            rendered_title = inline_markdown(title, source)
            output.append(f'<h{level} id="{heading_id}">{rendered_title}</h{level}>')
            if level in (2, 3, 4):
                headings.append({"level": level, "id": heading_id, "title": plain_text(rendered_title)})
            i += 1
            continue

        if re.match(r"^\s{0,3}(?:[-*_]\s*){3,}$", line):
            output.append('<hr class="sketch-rule">')
            i += 1
            continue

        if re.match(r"^\s{0,3}>", line):
            quote_lines = []
            while i < len(lines):
                current = lines[i]
                if re.match(r"^\s{0,3}>", current):
                    quote_lines.append(re.sub(r"^\s{0,3}> ?", "", current))
                    i += 1
                elif not current.strip() and i + 1 < len(lines) and re.match(r"^\s{0,3}>", lines[i + 1]):
                    quote_lines.append("")
                    i += 1
                else:
                    break
            output.append('<blockquote class="sketch-quote">')
            output.append(render_markdown("\n".join(quote_lines), source, headings))
            output.append("</blockquote>")
            continue

        if is_table_start(lines, i):
            headers = split_table_row(lines[i])
            align_cells = split_table_row(lines[i + 1])
            alignments = []
            for cell in align_cells:
                cell = cell.strip()
                alignments.append("center" if cell.startswith(":") and cell.endswith(":") else "right" if cell.endswith(":") else "left")
            i += 2
            rows = []
            while i < len(lines) and lines[i].strip() and "|" in lines[i]:
                rows.append(split_table_row(lines[i]))
                i += 1
            table = ['<div class="table-wrap"><table><thead><tr>']
            for column, title in enumerate(headers):
                alignment = alignments[column] if column < len(alignments) else "left"
                table.append(f'<th style="text-align:{alignment}">{inline_markdown(title, source)}</th>')
            table.append("</tr></thead><tbody>")
            for row in rows:
                table.append("<tr>")
                for column, cell in enumerate(row):
                    alignment = alignments[column] if column < len(alignments) else "left"
                    table.append(f'<td style="text-align:{alignment}">{inline_markdown(cell, source)}</td>')
                table.append("</tr>")
            table.append("</tbody></table></div>")
            output.append("".join(table))
            continue

        marker = list_marker(line)
        if marker:
            base_indent = len(marker.group(1).replace("\t", "    "))
            ordered = marker.group(2)[0].isdigit()
            tag = "ol" if ordered else "ul"
            items: list[str] = []
            while i < len(lines):
                current = lines[i]
                item_match = list_marker(current)
                if not item_match:
                    break
                indent = len(item_match.group(1).replace("\t", "    "))
                current_ordered = item_match.group(2)[0].isdigit()
                if indent != base_indent or current_ordered != ordered:
                    break
                item = item_match.group(3).strip()
                task = re.match(r"^\[([ xX])\]\s*(.*)$", item)
                if task:
                    checked = " checked" if task.group(1).lower() == "x" else ""
                    item = f'<input class="task-box" type="checkbox" disabled{checked}> ' + task.group(2)
                items.append("<li>" + inline_markdown(item, source) + "</li>")
                i += 1
                # Join indented continuation lines; source documents mostly use flat lists.
                while i < len(lines) and lines[i].strip() and not list_marker(lines[i]) and len(lines[i]) - len(lines[i].lstrip()) > base_indent:
                    continuation = lines[i].strip()
                    items[-1] = items[-1][:-5] + " " + inline_markdown(continuation, source) + "</li>"
                    i += 1
            output.append(f'<{tag} class="lesson-list">' + "".join(items) + f"</{tag}>")
            continue

        # A paragraph may span multiple source lines; stop at the next block element.
        paragraph = [line.rstrip()]
        i += 1
        while i < len(lines) and lines[i].strip() and not starts_block(i):
            paragraph.append(lines[i].rstrip())
            i += 1
        joined = " ".join(part.strip() for part in paragraph)
        joined = re.sub(r"\\\\\s*$", "<br>", joined)
        output.append(f'<p>{inline_markdown(joined, source)}</p>')

    return "\n".join(output)


def get_lead(markdown: str) -> str:
    lines = markdown.splitlines()
    for index, line in enumerate(lines):
        if line.startswith("# "):
            continue
        if line.strip():
            if line.lstrip().startswith("> "):
                quote_lines = []
                for current in lines[index:]:
                    if current.lstrip().startswith(">"):
                        quote_lines.append(re.sub(r"^\s*> ?", "", current))
                    elif not current.strip() and quote_lines:
                        break
                    else:
                        break
                return " ".join(part.strip() for part in quote_lines if part.strip())
            break
    return "Go 基础语法，一篇一篇学会。"


def render_sidebar(active_id: str = "") -> str:
    progress = (
        '<div class="progress-card sketch-card">'
        '<div class="progress-top"><span>学习进度</span><strong data-progress-label>0 / 10</strong></div>'
        '<div class="progress-track" aria-label="学习进度"><span data-progress-bar></span></div>'
        '<p>学完一篇，记得给自己一个小勾勾 ✓</p>'
        '</div>'
    )
    search = '<label class="search-field"><span class="sr-only">搜索章节</span><span class="search-icon" aria-hidden="true">⌕</span><input type="search" placeholder="找一篇 Go 知识…" data-search-lessons></label>'
    links = []
    for lesson in LESSONS:
        active = ' aria-current="page"' if lesson["id"] == active_id else ""
        links.append(
            f'<a class="course-link" href="{esc(lesson["page"])}" data-lesson-id="{esc(lesson["id"])}" data-title="{esc(lesson["name"])}"{active}>'
            f'<span class="course-number">{esc(lesson["id"])}</span><span class="course-name">{esc(lesson["name"])}</span>'
            f'<span class="course-check" aria-hidden="true">✓</span></a>'
        )
    return (
        '<aside class="side-col" aria-label="课程导航">'
        f'{progress}{search}<nav class="course-nav"><div class="nav-heading">学习路线 <span>✎</span></div>{"".join(links)}</nav>'
        '<a class="code-index-link" href="../../codes/教学/README.md">📎 全部示例代码</a>'
        '</aside>'
    )


def render_header() -> str:
    return (
        '<a class="skip-link" href="#main-content">跳到正文</a>'
        '<header class="topbar"><div class="topbar-inner">'
        '<a class="brand" href="index.html" aria-label="Go 代码探险队首页">'
        '<span class="brand-mark" aria-hidden="true"><span>go</span><i>✦</i></span>'
        '<span class="brand-copy"><strong>Go 代码探险队</strong><small>一支笔 · 一段代码 · 一个知识点</small></span></a>'
        '<nav class="top-links" aria-label="主要导航">'
        '<a href="index.html">学习地图</a><a href="00-roadmap.html">完整路线文稿</a><a href="../../codes/教学/README.md">示例代码</a></nav>'
        '<div class="top-actions"><button class="tool-button focus-button" type="button" data-action="focus">专注阅读</button>'
        '<button class="tool-button menu-button" type="button" data-action="menu" aria-expanded="false" aria-label="打开课程导航">☰</button></div>'
        '</div></header>'
    )


def render_code_groups(lesson: dict) -> str:
    groups = []
    for group in lesson.get("code_groups", []):
        files = []
        for relative in group["files"]:
            path = CODE_ROOT / relative
            if not path.is_file():
                raise FileNotFoundError(f"配套代码缺失：{path}")
            code = path.read_text(encoding="utf-8").rstrip("\n")
            link = "../../codes/教学/" + "/".join(quote(part, safe="-._~") for part in Path(relative).parts)
            files.append(
                '<details class="source-file">'
                f'<summary><span class="source-path">{esc(relative)}</span><span class="source-open">展开源码 <b>＋</b></span></summary>'
                f'<div class="source-actions"><a href="{esc(link)}" download>下载原文件 ↗</a><span>代码已嵌入本页，可离线查看</span></div>'
                f'{render_code_block(code, "go")}</details>'
            )
        command = group.get("run", "")
        groups.append(
            '<section class="code-group">'
            f'<h3>{esc(group["title"])}</h3>'
            + (f'<div class="run-command"><span aria-hidden="true">▶</span><code>cd 00-基础语法/codes/教学 &amp;&amp; {esc(command)}</code></div>' if command else "")
            + "".join(files)
            + "</section>"
        )
    if not groups:
        return (
            '<section class="code-library" id="companion-code"><div class="section-stamp">🗺️ 学习地图</div>'
            '<h2>把路线图当作你的导航</h2><p>从 01 开始，一篇一篇向前走；每篇文稿下面都有可以直接复制运行的 Go 源码。</p>'
            '<a class="button button-primary" href="01-setup.html">从第 01 篇出发 <span>→</span></a></section>'
        )
    return (
        '<section class="code-library" id="companion-code">'
        '<div class="section-stamp">🧰 本章工具箱</div><h2>配套源码（已嵌入）</h2>'
        '<p class="code-library-intro">下面的源码来自仓库中的 <code>codes/教学</code>，展开即可复制；页面不需要联网加载代码。</p>'
        + "".join(groups)
        + '<a class="repo-code-link" href="../../codes/教学/README.md">看完整代码目录与运行说明 ↗</a>'
        + '</section>'
    )


def toc_markup(headings: list[dict]) -> str:
    if not headings:
        return '<p class="toc-empty">本页暂无小节导航。</p>'
    items = []
    for item in headings:
        if item["level"] == 4:
            continue
        klass = "toc-sub" if item["level"] == 3 else ""
        items.append(f'<li class="{klass}"><a href="#{esc(item["id"])}">{esc(item["title"])}</a></li>')
    return '<ul class="toc-list">' + "".join(items) + "</ul>"


def render_lesson_page(lesson: dict) -> str:
    source = SOURCE_DIR / lesson["file"]
    markdown = source.read_text(encoding="utf-8")
    lines = markdown.splitlines()
    if lines and lines[0].startswith("# "):
        lines = lines[1:]
    while lines and not lines[0].strip():
        lines = lines[1:]
    # The first blockquote is the learning objective/intro already shown in the hero.
    if lines and lines[0].lstrip().startswith(">"):
        while lines and lines[0].lstrip().startswith(">"):
            lines = lines[1:]
        while lines and not lines[0].strip():
            lines = lines[1:]
    markdown = "\n".join(lines)
    heading_items: list[dict] = []
    article = render_markdown(markdown, source, heading_items)
    lead = inline_markdown(get_lead(source.read_text(encoding="utf-8")), source)
    course_id = lesson["id"]
    page_title = f'{course_id} {lesson["name"]}'
    is_overview = course_id == "00"
    source_href = "../" + quote(lesson["file"], safe="-._~")
    prev_next = []
    if not is_overview:
        index = next(i for i, item in enumerate(LESSONS) if item["id"] == course_id)
        if index > 1:
            prev = LESSONS[index - 1]
            prev_next.append(f'<a class="previous-lesson" href="{esc(prev["page"])}"><span>← 上一篇</span><strong>{esc(prev["id"])} {esc(prev["name"])}</strong></a>')
        if index + 1 < len(LESSONS):
            nxt = LESSONS[index + 1]
            prev_next.append(f'<a class="next-lesson" href="{esc(nxt["page"])}"><span>下一站 →</span><strong>{esc(nxt["id"])} {esc(nxt["name"])}</strong></a>')
    complete_control = "" if is_overview else (
        f'<button class="complete-button" type="button" data-action="complete" data-lesson="{esc(course_id)}">'
        '<span class="complete-icon">✓</span><span data-complete-label>我学会了，打个勾</span></button>'
    )
    answer = f'<p class="recall-answer" hidden>{inline_markdown(lesson["hint"], source)}</p>'
    return f'''<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="theme-color" content="#fff8e9">
  <meta name="description" content="{esc(lesson['tagline'])}">
  <title>{esc(page_title)} · Go 代码探险队</title>
  <link rel="stylesheet" href="assets/style.css">
</head>
<body class="lesson-page" data-page="lesson" data-lesson="{esc(course_id)}">
{render_header()}
<div class="page-grid">
  {render_sidebar(course_id)}
  <main class="lesson-main" id="main-content">
    <div class="breadcrumbs"><a href="index.html">学习地图</a><span>／</span><span>{esc(lesson['phase'])}</span><span>／</span><strong>{esc(course_id)}</strong></div>
    <section class="chapter-hero sketch-card">
      <div class="chapter-kicker"><span class="chapter-sticker">任务 {esc(course_id)}</span><span>{esc(lesson['phase'])}</span><span class="hero-spark" aria-hidden="true">✦</span></div>
      <h1>{esc(page_title)}</h1>
      <p class="chapter-lead">{lead}</p>
      <div class="chapter-meta"><span>📝 一篇一练</span><span>🧩 小步掌握</span><a href="{esc(source_href)}">查看 Markdown 原稿 ↗</a></div>
      <svg class="hero-doodle-mini" viewBox="0 0 100 90" aria-hidden="true"><path d="M12 65 Q28 42 41 58 T70 44" fill="none" stroke="#287d68" stroke-width="3" stroke-linecap="round" stroke-dasharray="2 7"/><path d="M57 28 L77 9 L91 23 L71 43 Z" fill="#ffd46c" stroke="#263238" stroke-width="2.4" stroke-linejoin="round"/><path d="M57 28 L71 43 L67 52 L52 36 Z" fill="#f58a70" stroke="#263238" stroke-width="2.4"/><path d="M52 36 L47 54 L67 52" fill="#fff8e9" stroke="#263238" stroke-width="2.4" stroke-linejoin="round"/><path d="M21 75 Q33 82 42 74" fill="none" stroke="#e78361" stroke-width="2" stroke-linecap="round"/></svg>
    </section>
    <section class="memory-note sketch-card" aria-label="记忆提示">
      <div class="memory-title"><span aria-hidden="true">✎</span> 一句话记住</div>
      <p class="memory-line">{inline_markdown(lesson['memory'], source)}</p>
      <div class="recall-row"><span class="recall-badge">🧠 30 秒复述</span><span class="recall-question">{esc(lesson['prompt'])}</span><button type="button" class="hint-button" data-action="reveal">翻开提示 ↗</button></div>
      {answer}
    </section>
    {complete_control}
    <article class="markdown-body" id="article">
      {article}
    </article>
    {render_code_groups(lesson)}
    <nav class="lesson-pager" aria-label="篇章翻页">{"".join(prev_next) if prev_next else '<a class="previous-lesson" href="index.html"><span>← 回到地图</span><strong>选一篇开始探险</strong></a>'}</nav>
    <footer class="page-footer"><span>✿ 边看边敲，比只看更牢。</span><span>源稿：<a href="{esc(source_href)}">{esc(lesson['file'])}</a></span></footer>
  </main>
  <aside class="toc-col" aria-label="本页导航"><div class="toc-card sketch-card"><div class="toc-title">本页小地图 <span>⌁</span></div><nav id="toc-nav">{toc_markup(heading_items)}</nav><div class="toc-tip">读完一小节，就自己讲一遍。<span>↙</span></div></div></aside>
</div>
<div class="mobile-scrim" data-action="close-menu" aria-hidden="true"></div>
<script src="assets/app.js" defer></script>
</body>
</html>'''


def render_index_page() -> str:
    phase_cards = []
    for index, (phase_id, name, description, lesson_ids, icon) in enumerate(PHASES, start=1):
        cards = []
        for lesson_id in lesson_ids:
            lesson = LESSON_BY_ID[lesson_id]
            cards.append(
                f'<a class="route-lesson" href="{esc(lesson["page"])}" data-lesson-id="{esc(lesson_id)}" data-title="{esc(lesson["name"])}">'
                f'<span class="route-number">{esc(lesson_id)}</span><span class="route-lesson-name">{esc(lesson["name"])}</span><span class="route-check" aria-hidden="true">✓</span></a>'
            )
        phase_cards.append(
            f'<article class="phase-card phase-{index} sketch-card"><div class="phase-head"><span class="phase-icon" aria-hidden="true">{icon}</span><span class="phase-index">STAGE {index:02d}</span></div>'
            f'<h3>{esc(name)}</h3><p>{esc(description)}</p><div class="route-list">{"".join(cards)}</div><span class="phase-scribble" aria-hidden="true">{("↘", "↗", "↘", "✦")[index - 1]}</span></article>'
        )
    first = LESSON_BY_ID["01"]
    return f'''<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="theme-color" content="#fff8e9">
  <meta name="description" content="Go 入门手绘学习站：教学文稿、互动记忆卡与可复制的 Go 示例代码。">
  <title>Go 代码探险队 · 手绘学习地图</title>
  <link rel="stylesheet" href="assets/style.css">
</head>
<body class="home-page" data-page="home">
{render_header()}
<main id="main-content" class="home-wrap">
  <section class="home-hero sketch-card">
    <div class="hero-copy">
      <div class="eyebrow"><span class="eyebrow-dot"></span> GO 入门 · 手绘学习册</div>
      <h1>把 Go 学习<br><em>画成一场小冒险</em></h1>
      <p>从第一行 <code>fmt.Println</code> 出发，沿着“能写 → 写对 → 写好 → 写快”的路线，边读边敲，逐篇通关。</p>
      <div class="hero-actions"><a class="button button-primary" href="{esc(first['page'])}">从第 01 篇开始 <span>→</span></a><a class="button button-quiet" href="00-roadmap.html">先看完整路线</a></div>
      <div class="hero-facts"><span><b>10</b> 篇循序渐进</span><span><b>20+</b> 个代码练习包</span><span><b>0</b> 个外部依赖</span></div>
    </div>
    <div class="hero-art" aria-label="一位拿着铅笔的小小编程探险家">
      <div class="art-sun"></div><div class="art-caption">今天也要<br>敲一点点 ✎</div>
      <svg viewBox="0 0 430 330" class="explorer-svg" role="img" aria-label="手绘书本、路线和星星">
        <path d="M39 231 C83 166 96 286 153 209 S231 183 265 218 S326 265 383 163" fill="none" stroke="#368875" stroke-width="4" stroke-linecap="round" stroke-dasharray="2 12"/>
        <path d="M74 166 L98 151 L126 196 L101 211 Z" fill="#ffd871" stroke="#263238" stroke-width="3" stroke-linejoin="round"/>
        <path d="M153 220 C155 173 192 149 236 166 L253 172 L253 242 C215 225 185 226 153 249 Z" fill="#fffdf5" stroke="#263238" stroke-width="3" stroke-linejoin="round"/>
        <path d="M253 172 C293 150 327 165 345 188 L345 252 C316 228 283 225 253 242 Z" fill="#f6a17f" stroke="#263238" stroke-width="3" stroke-linejoin="round"/>
        <path d="M202 189 Q221 185 237 194 M201 205 Q218 201 238 209 M274 190 Q300 177 321 190 M274 207 Q300 195 321 208" fill="none" stroke="#6c8c81" stroke-width="3" stroke-linecap="round"/>
        <path d="M280 95 L314 50 L330 64 L296 109 Z" fill="#ffd871" stroke="#263238" stroke-width="3" stroke-linejoin="round"/>
        <path d="M280 95 L296 109 L284 126 L272 113 Z" fill="#ee8469" stroke="#263238" stroke-width="3" stroke-linejoin="round"/>
        <path d="M272 113 L264 135 L284 126 Z" fill="#fffdf5" stroke="#263238" stroke-width="3" stroke-linejoin="round"/>
        <path d="M105 82 l6 15 16 2 -12 10 3 16 -13 -8 -14 8 4 -15 -11 -11 16 -1z" fill="#f2c44e" stroke="#263238" stroke-width="2.5" stroke-linejoin="round"/>
        <path d="M366 95 l4 10 11 1 -8 7 2 11 -9 -6 -10 6 3 -11 -8 -7 11 -1z" fill="#72b5a0" stroke="#263238" stroke-width="2.3" stroke-linejoin="round"/>
        <path d="M350 274 Q370 257 391 268" fill="none" stroke="#ee8469" stroke-width="3" stroke-linecap="round"/>
        <circle cx="64" cy="257" r="4" fill="#ee8469"/><circle cx="130" cy="133" r="3" fill="#4c9c84"/>
      </svg>
      <div class="art-sticker">读 · 写 · 跑 · 讲</div>
    </div>
    <span class="hero-doodle-arrow" aria-hidden="true">先从这里出发 ↘</span>
  </section>

  <section class="route-section" id="route-map">
    <div class="section-heading"><div><span class="section-overline">FOLLOW THE DOODLE</span><h2>你的 Go 探险路线</h2><p>不用一口气学完。每次通关一站，知识就多一块拼图。</p></div><label class="search-field route-search"><span aria-hidden="true">⌕</span><input type="search" placeholder="搜章节：比如切片、并发…" data-search-cards><span class="sr-only">搜索课程</span></label></div>
    <div class="phase-grid">{"".join(phase_cards)}</div>
    <p class="search-empty" data-search-empty hidden>没有找到这篇，试试“函数”或“JSON”。</p>
    <div class="route-note"><span class="note-pin" aria-hidden="true">✦</span><span>小提示：<strong>每篇都有配套 Go 源码</strong>。读完先猜输出，再展开代码运行，记忆会更牢。</span><a href="00-roadmap.html">打开完整路线文稿 ↗</a></div>
  </section>

  <section class="practice-section">
    <div class="practice-intro"><span class="section-overline">A TINY BRAIN BREAK</span><h2>抽一张记忆卡</h2><p>先想一想，再点开答案。<br>主动回忆，比重读三遍更管用。</p><svg viewBox="0 0 130 90" aria-hidden="true"><path d="M12 68 Q35 14 68 50 T119 22" fill="none" stroke="#ed8469" stroke-width="3" stroke-dasharray="2 8" stroke-linecap="round"/><path d="M103 13 l7 12 14 2 -10 9 2 14 -13 -7 -12 7 3 -14 -10 -9 14 -2z" fill="#ffd871" stroke="#263238" stroke-width="2"/></svg></div>
    <div class="flashcard sketch-card" data-flashcard>
      <div class="flashcard-top"><span>🃏 快速回忆</span><span data-flash-count>01 / 05</span></div>
      <p class="flash-question" data-flash-question>Go 的数组和切片，哪个长度固定？</p>
      <p class="flash-answer" data-flash-answer hidden>数组长度写在类型里、固定不变；切片是对底层数组的一扇可变窗口。</p>
      <div class="flash-actions"><button class="button button-primary" type="button" data-action="flip-flash">翻开答案</button><button class="button button-quiet" type="button" data-action="next-flash">换一张 ↻</button></div>
    </div>
  </section>

  <section class="study-steps">
    <div class="section-heading"><div><span class="section-overline">HOW TO LEARN</span><h2>每一站，用三步学会</h2></div><a class="all-code-link" href="../../codes/教学/README.md">打开全部示例代码 ↗</a></div>
    <div class="step-grid">
      <article class="step-card step-read"><span class="step-num">01</span><span class="step-icon">👀</span><h3>先读一遍</h3><p>看懂概念和手绘记忆句，不急着背。</p><span class="step-doodle">轻轻划重点</span></article>
      <article class="step-card step-run"><span class="step-num">02</span><span class="step-icon">⌨️</span><h3>亲手跑代码</h3><p>展开本章源码，复制到终端运行，留意输出。</p><span class="step-doodle">别只看哦！</span></article>
      <article class="step-card step-tell"><span class="step-num">03</span><span class="step-icon">🗣️</span><h3>讲给别人听</h3><p>合上页面，用自己的话复述，再给本章打勾。</p><span class="step-doodle">真的学会了</span></article>
    </div>
  </section>
  <footer class="home-footer"><span class="footer-scribble">慢慢写，都会的。</span><span>内容来源：<a href="../00-索引与学习路线.md">教学文稿 Markdown</a> · HTML 页面可离线阅读</span></footer>
</main>
<script src="assets/app.js" defer></script>
</body>
</html>'''


def main() -> None:
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    (OUTPUT_DIR / "assets").mkdir(exist_ok=True)
    home = render_index_page()
    (OUTPUT_DIR / "index.html").write_text(home, encoding="utf-8")
    for lesson in LESSONS:
        page = render_lesson_page(lesson)
        (OUTPUT_DIR / lesson["page"]).write_text(page, encoding="utf-8")
    total = sum(path.stat().st_size for path in OUTPUT_DIR.glob("*.html"))
    print(f"Built {len(LESSONS) + 1} HTML pages in {OUTPUT_DIR.relative_to(REPO_ROOT)} ({total:,} bytes)")


if __name__ == "__main__":
    main()
