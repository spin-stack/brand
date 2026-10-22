# spin-stack / brand

The colors, type, layout and mark of Spinbox, written once, checked, and made into the files each
place that draws Spinbox reads: the website (Starlight), the dashboard (Tailwind, shadcn's
variables) and anything else (`dist/tokens.json`).

**[See it](https://spin-stack.github.io/brand/)**: `index.html` is the laboratory.
It draws the site's landing and a docs page, the dashboard's workspaces with their states, a
dialog and a toast; the app as it is, in screenshots spin's own suite takes; the mark's guide -
sizes, clear space, the lockup, the grounds it goes on and what it is not; the palette by what
each color is for; and the same page at a phone's and a tablet's width - in dark or light and
through protanopia, deuteranopia or tritanopia, with every pair the palette is held to.

| Source | What it makes |
|---|---|
| `palette.toml` | `dist/`: `tokens.css` (every role as `--spin-*`), `starlight.css` (the website's `--sl-color-*`), `dashboard.css` (shadcn's HSL variables, light on `:root`, dark on `.dark`), `tokens.json`, and `perception.js` and `simulate.svg` for the preview |
| spin's `dashboard/e2e/site.spec.ts` | `shots/`: the app as it is, screenshots spin's dashboard suite takes of its e2e stack, with the release and commit of the spin tree that took them (`task screenshots SHOTS=<dir>`); the website copies them from here |
| `assets/avatar.png` | `logos/`: the stack of disks cut from the organization's avatar, without the wordmark: the tile on its own navy (512, 256; the favicon at 64 and 32) for a dark ground, and clear - the ground taken out - (512, 256) for a light one |

```
task            # palette and mark: check, and write dist/ and logos/
task report     # every pair, its contrast or its distance in each view
task preview    # http://127.0.0.1:4330/
task check      # what CI runs: fails on a pair below its floor, or a stale dist/ or logos/
```

## What holds the palette

`cmd/palette` checks every pair that carries meaning, in both modes, before it writes anything:

- **Contrast**, WCAG 2.2: the product's text at 7:1 on every surface, secondary text, links,
  labels and status text at 4.5:1, an input's edge and the focus ring at 3:1.
- **Distance**, CIEDE2000 (checked against Sharma's test data): two surfaces that meet, a border
  against what it bounds, a selected and a hovered row against a plain one, and every status
  against every other and against the accent - as drawn, on an office's IPS panel, on the same
  panel at 30% backlight, and through each dichromacy (Machado, Oliveira and Fernandes 2009,
  severity 1.0), protan and deutan also on the IPS panel. The floors are in `checks.go`, each with
  its reason; they follow spin-stack's sibling CONSTRUCT.
- **No meaning by hue alone**: a status carries a shape too (the preview's states: a dot, a
  diamond, a square), since a floor for colour-vision deficiencies is a floor, not comfort.

Type and layout are tokens too: IBM Plex Sans and Mono (`fonts/`, OFL), a 16 px reading size
(below it iOS zooms into a focused input) and 14 px for the dashboard's tables, a 72ch measure,
24 px for any pointer target and 44 px on a coarse pointer (WCAG 2.2, 2.5.8 and 2.5.5).

## Where it goes

Each consumer copies the file made for it and holds the copy to this repository's, as the
website's refgen holds its reference to spin: a stale copy fails that consumer's check, not this
one's. The mapping from roles to each consumer's variables is `cmd/palette/outputs.go`.

## License

Apache License 2.0, as spin. The fonts are IBM Plex under the SIL Open Font License
(`fonts/LICENSE-IBM-Plex.txt`).
