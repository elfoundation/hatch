# Hatch brand usage

> One-pager. Read this before you put the wordmark anywhere.

## Files in this kit

```
brand/
├── wordmark/
│   ├── wordmark.svg              # master wordmark (text-only, with embedded font)
│   ├── wordmark-dark.svg         # dark variant (paper on transparent)
│   ├── wordmark-on-light.svg     # paper-on-warm-white background
│   ├── wordmark-on-dark.svg      # paper-on-ink background
│   └── wordmark-*.png            # 200x60, 400x120, 800x240, 1200x360
├── mark/
│   ├── mark.svg                  # the icon alone (no text)
│   ├── mark-on-light.svg
│   ├── mark-on-dark.svg
│   └── mark-{16,32,48,64,128,192,256,512,1024}.png
├── favicon/
│   ├── favicon.ico               # 16+32+48 multi-size ico
│   └── favicon-{16,32,48,192,512}.png
├── banner/
│   ├── readme-banner.png         # 1280x640
│   └── readme-banner@2x.png      # 2560x1280
└── og/
    ├── og.png                    # 1200x630
    └── og@2x.png                 # 2400x1260
```

## What goes where

| Surface                                  | Use                              | File                                   |
| ---------------------------------------- | -------------------------------- | -------------------------------------- |
| Favicon (browser tab, GitHub avatar)     | The **mark**, not the wordmark   | `brand/favicon/favicon.ico`            |
| Apple touch icon / PWA icon               | The mark, 512px                  | `brand/favicon/favicon-512.png`        |
| README header                            | The **wordmark** (200x60+)       | `brand/wordmark/wordmark-200x60.png`   |
| GitHub social preview / blog / link card | The **OG image** (1200x630)      | `brand/og/og.png`                      |
| Big surfaces, slides, README top         | The **README banner** (1280x640) | `brand/banner/readme-banner.png`       |

For the favicon, always use `mark.svg` or one of the favicon PNGs — never the full
wordmark. The wordmark does not read at 16x16.

## Clear space

Treat the mark as a square and the wordmark as the mark + the word. Clear space
is **1× the cap-height of the "h" in "hatch"** on every side.

- Wordmark: at least one "h"-height of empty space on top, bottom, left, and right
- Mark: at least one "h"-height on every side

```
   ┌──────────────────────────────┐
   │                              │  ← 1× clear space
   │   ┌──┐  hatch                │
   │   │  │                       │
   │   └──┘                       │
   │                              │  ← 1× clear space
   └──────────────────────────────┘
```

## Minimum size

- Wordmark: **120px wide** is the minimum. Below that, switch to the mark.
- Mark: **16px** is the minimum (we ship a tuned 16px favicon).

## Color

The brand has **one active color, one neutral, and one accent**. That's it.

| Token    | Hex       | Use                                               |
| -------- | --------- | ------------------------------------------------- |
| Ink      | `#0A0A0B` | Wordmark, primary type, dark surfaces             |
| Paper    | `#FAFAF9` | Light surfaces, type on dark                      |
| Accent   | `#F59E0B` | The "captured" signal — one place per composition |

Neutral scale (for docs and UIs, not the wordmark): `#71717A` `#A1A1AA` `#D4D4D8` `#E4E4E7` `#F4F4F5`.

**Don't pick colors outside the kit.** If you need another color, ask the Head of
Marketing; do not invent one.

## Light and dark mode

The wordmark and the mark are **single-color**. Pick the right variant for the
surface — never change the color of the wordmark to fit a theme.

- **On a light surface** (≥ 4.5:1 contrast): use `wordmark.svg` (ink) or `mark.svg`
- **On a dark surface**: use `wordmark-dark.svg` (paper) or `mark-dark.svg`
- **Need a background tile?** Use `wordmark-on-light.svg` or `wordmark-on-dark.svg`

For GitHub READMEs, GitHub auto-switches images in dark mode only if you use
the `#gh-dark-mode-only` / `#gh-light-mode-only` MediaFragments trick:

```markdown
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="brand/wordmark/wordmark-dark.svg">
  <img src="brand/wordmark/wordmark.svg" alt="Hatch" width="200">
</picture>
```

## Typography (display + body)

The wordmark uses **Inter Bold**. The brand voice guide and READMEs use:

- **Display / headings:** Inter (Bold or SemiBold)
- **Code / mono:** JetBrains Mono (Regular or Medium)

Both are OFL-licensed and on the system. Don't substitute a different sans or mono
in brand surfaces without sign-off.

## Do not

- **Don't stretch.** Always use the mark/wordmark at its native aspect ratio. If
  you need it bigger, scale proportionally.
- **Don't recolor.** The wordmark is ink or paper — period. No brand-color,
  gradient, rainbow, or AI-painted versions.
- **Don't add effects.** No drop shadows, no glows, no outlines, no bevel, no
  inner shadows. The mark is a flat shape.
- **Don't rotate.** The hatch opens at the top. Don't tilt the mark.
- **Don't rearrange the wordmark.** The mark always sits to the **left** of the
  wordmark, with a 0.22×-mark-width gap. No stacking, no right-alignment.
- **Don't put the wordmark on a busy image.** The wordmark needs contrast. On a
  photo or a complex background, use a solid color tile behind it.
- **Don't use the wordmark in body copy.** The wordmark is for headers, banners,
  and brand surfaces only. Body type stays in Inter.
- **Don't use the favicon as a profile picture.** Use the 512px PNG (`mark-512.png`)
  for that, and round the corners if you want — but the square mark is the
  canonical shape.

## Licensing & attribution

- **Inter** — SIL Open Font License 1.1 (Inter by Rasmus Andersson, rsms.me)
- **JetBrains Mono** — SIL Open Font License 1.1 (JetBrains)
- The mark and wordmark are the property of El Foundation. Use them to talk about
  Hatch. Do not register them as a trademark or use them to sell a competing
  product.

## Need something new?

If a new surface needs a brand asset that isn't in the kit (a slide template, a
business card, a T-shirt, an icon variant), open an issue assigned to the Brand
Designer. Do not stretch, recolor, or rearrange this kit to make it work.
