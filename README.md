# nlm_clear_by_more_money 💸

> When NotebookLM mangles your Chinese text, throw money at the problem.

## The Problem

You've probably been there: You upload a beautiful PDF to NotebookLM, excited to generate that fancy AI podcast or summary. Then you see the output — your Chinese characters look like they went through a blender. 你好 becomes 祢女子. Chaos ensues.

The internet offers many "solutions":
- **OCR-based approaches**: Re-extract text, rebuild slides, pray it works
- **Manual fixes**: Hire an intern, or become one yourself
- **Acceptance**: Just pretend those characters are "artistic"

## Our Solution: Violence (Financial)

This project takes a refreshingly honest approach: **just pay Gemini to fix it**.

No clever algorithms. No sophisticated OCR pipelines. Just raw, unfiltered API calls at 2K resolution, burning through tokens like there's no tomorrow.

```
PDF → Images (300 DPI) → Gemini "please fix this" → PPTX
```

It's not elegant. It's not cheap. But it works.

## Features

- 🔥 **Parallel processing** — Burn money faster with concurrent API calls
- 📈 **Configurable resolution** — 1K, 2K, or 4K (for when you really hate your wallet)
- 🎯 **Pure Go** — Single binary, no dependencies, no excuses
- 📦 **Cross-platform** — Windows, macOS, Linux

## Installation

### From Source

```bash
git clone https://github.com/eternnoir/nlm_clear_by_more_money.git
cd nlm_clear_by_more_money
make build
```

### Pre-built Binaries

Download from [Releases](https://github.com/eternnoir/nlm_clear_by_more_money/releases) page.

## Usage

```bash
# Set your API key (the money pipeline)
export GEMINI_API_KEY="your-api-key"

# Basic usage
./nlm_clear_by_more_money input.pdf

# Output: input.pptx
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-size` | Image resolution (1K, 2K, 4K) | 2K |
| `-parallel` | Concurrent API requests | 5 |
| `-prompt` | Custom enhancement prompt | Built-in |
| `-temp` | Generation temperature (0.0-1.0) | 0.6 |
| `-no-enhance` | Skip Gemini (for testing) | false |

### Examples

```bash
# Go faster, spend more
./nlm_clear_by_more_money -parallel 10 input.pdf

# Maximum quality, maximum regret
./nlm_clear_by_more_money -size 4K -parallel 10 input.pdf

# Test without API calls (free!)
./nlm_clear_by_more_money -no-enhance input.pdf

# Custom prompt
./nlm_clear_by_more_money -prompt "Fix the text, keep the vibes" input.pdf
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GEMINI_API_KEY` | Your Gemini API key (required) |
| `IMAGE_SIZE` | Default image size |
| `ENHANCE_PROMPT` | Default enhancement prompt |

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make all

# Build specific platform
make build-darwin-arm64
make build-darwin-amd64
make build-linux-amd64
make build-windows-amd64

# Clean build artifacts
make clean
```

## How It Works

1. **PDF → Images**: Uses [go-pdfium](https://github.com/klippa-app/go-pdfium) (WASM) to render each page at 300 DPI
2. **Images → Enhanced Images**: Sends to Gemini API with a polite request to fix blurry text
3. **Enhanced Images → PPTX**: Packages everything into a valid PPTX file

The PPTX builder is pure Go — no external dependencies, no Office installation required.

## FAQ

**Q: Is this actually a good solution?**
A: It's *a* solution. Whether it's good depends on how much you value your sanity vs. your money.

**Q: Why not just use OCR?**
A: OCR rebuilds text. We rebuild images. Different philosophy, same desperation.

**Q: The output still has some broken characters.**
A: Try 4K resolution. Or accept that some things in life are beyond repair.

**Q: Can I use this for languages other than Chinese?**
A: Yes! Blurry text is a universal problem. Use `-prompt` to customize.

## License

MIT — Use it, fork it, blame it.

## Acknowledgments

- [go-pdfium](https://github.com/klippa-app/go-pdfium) — PDF rendering without CGO
- [Google Gemini](https://ai.google.dev/) — The expensive magic
- NotebookLM — For creating the problem this solves

---

*Remember: Every token spent is a character saved.* 💸
