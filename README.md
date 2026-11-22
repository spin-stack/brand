# spin-stack / brand

The colors, type, layout and mark of Spinbox, written once, checked, and made into the files each
place that draws Spinbox reads: the website (Starlight), the dashboard (Tailwind, shadcn's
variables), the agent client the dashboard frames (the T3 Code fork's variables) and anything
else (`dist/tokens.json`).

**[See it](https://spin-stack.github.io/brand/)**: `index.html` is the laboratory.
It draws the site's landing and a docs page, the dashboard's workspaces with their states, a
dialog and a toast; where the app as it is is shown; the mark's guide -
sizes, clear space, the lockup, the grounds it goes on and what it is not; the type scale in both
faces and the three weights; the palette by what each color is for; and the same page at a
phone's and a tablet's width - in dark or light and through protanopia, deuteranopia or
tritanopia, with every pair the palette is held to.

| Source | What it makes |
|---|---|
| `palette.toml` | `dist/`: `tokens.css` (every role as `--spin-*`), `starlight.css` (the website's `--sl-color-*`), `dashboard.css` (shadcn's HSL variables, light on `:root`, dark on `.dark`, and the dashboard's radius, scrim, dense type scale, heading tracking, motion and a terminal's black surface and sixteen ANSI colours as `--terminal-*`), `t3code.css` (the agent client's own variables, the roots its theme derives the rest from, on `html:root` and `html.dark:root`, with the dashboard's corner and the faces), `tokens.json`, and `perception.js` and `simulate.svg` for the preview |
| spin's `dashboard/e2e/catalog.spec.ts` | `screens/`: every route, tab, dialog and menu of the app in both themes at 1440 and 390 px, and what each draws that this palette does not have - a colour, a face, a size, a radius, contrast, overflow, a target too small (`task screens SCREENS=<dir>`). The catalog and the audit are committed; the images, hundreds a release, are an archive on a release of this repository that `screens.json` names and Pages unpacks. [See them](https://spin-stack.github.io/brand/screens.html) |
| `assets/mark.png` | `logos/`: the stack of disks, cut from the mark as it was drawn (1024 px on a clear ground): the tile on the brand's navy (512, 256; the touch icon at 180, the favicon at 64 and 32) for a dark ground, and clear (512, 256) for any ground: the drawing has nothing under the stack, so `mark-dark-*` (512, 256), the name the dark page asks for, is the clear mark. Each is the drawing reduced; a size it would have to enlarge fails |

```
go tool task            # palette and mark: check, and write dist/ and logos/
go tool task report     # every pair, its contrast or its distance in each view
go tool task preview    # http://127.0.0.1:4330/
go tool task check      # what CI runs: fails on a pair below its floor, or a stale dist/ or logos/
go tool task consumers  # which brand each consumer is on: its pin, its copies, its faces
go tool task agent-client T3CODE=<checkout>  # the fork's variable names, which t3code.css is held to
go tool task release    # a person's: name main's head v<YYYYMMDD>.<NN>, once Gate has passed on it
```

`go tool task` is [go-tools](https://github.com/spin-stack/go-tools)' task command, the one tool
`go.mod` takes: it runs the Task `versions.yaml` pins, the version CI runs, with go-tools' commands
on its PATH. `task check` also fails on a workflow step that is not one task (`taskgate`), on a
package install that skips Socket Firewall (`sfw check`), and on a face an output or a page names
that `fonts/` does not ship (`cmd/palette`'s tests).

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
- **Charts** (`[chart]`, `tokens.json`'s `chart`): eight series in a fixed order and three
  threshold statuses, one set for both modes because Grafana draws one dashboard JSON in either
  theme. Each holds 3:1 on dark's page and on light's, and each series is held apart from the
  next, and each status from the others, at the meaning tier.
- **A trace's services** (`[trace]`, `tokens.json`'s `trace`, `--trace-service-1` to `-8` in
  `dashboard.css`): eight for each theme, which the dashboard picks by a service name's hash.
  Each holds 4.5:1 on the page and 3:1 on a selected row, every pair is held apart, and each is
  held from `error` and from red, which are a failed span's.
- **No meaning by hue alone**: a status carries a shape too (the preview's states: a dot, a
  diamond, a square), since a floor for colour-vision deficiencies is a floor, not comfort.

Type and layout are tokens too: Inter and JetBrains Mono (`fonts/`, OFL) at three weights that
say rank (600 a page's title and a row's name, 500 a button and a table's header, 400 the rest;
`--spin-weight-*`), a 16 px reading size (below it iOS zooms into a focused input) and 13 px for
the dashboard's tables, a 72ch measure, 24 px for any pointer target and 44 px on a coarse pointer
(WCAG 2.2, 2.5.8 and 2.5.5).
Sizes are variables in every output, so that a consumer has no number of its own to drift from:
a control's three heights (28, 32 and 36 px, `--spin-control-*`), a list's row (36 px,
`--spin-row`), the radius, and the two sidebars, the docs' in rem (`--spin-sidebar`, and
Starlight's `--sl-sidebar-width`) and the app's in px (`--spin-sidebar-app`). Three lines, by what
each must be seen against: `hairline` separates on the page and is held against the page alone,
`border` is a line between two surfaces and a secondary button's outline (`--spin-control-border`,
`--control-border`), and `border_strong` is an input's edge at 3:1. `palette.toml` says the rules
that are behaviour: one filled action per view, a list's row as a plane under the pointer with
its actions drawn only then, an icon alone only with its caption.
Motion is four durations and one easing (`--spin-duration-*`, `--spin-ease`): a movement says
what something does or that it answered a hand, runs once, and is its end state under reduced
motion.

## Where it goes

[`consumers.toml`](consumers.toml) is every place that draws Spinbox: its repository, the files
it copies and where it keeps them, where its pin is, and the `package.json` its faces are
installed from. A place not in it is not a consumer, and `go test` fails on an output in `dist/`
that no row copies. The mapping from roles to each consumer's variables is
`cmd/palette/outputs.go`.

Each consumer pins a commit of this repository in its `versions.yaml` and holds its copies to
that commit in its own check: a stale copy fails there, not here. What this repository can say is
who is on what:

```
go tool task consumers                                    # each consumer, read through gh
go tool task consumers -- -checkout spin-stack/spin=../spin   # one of them from a checkout
```

It prints, for each, the commit and the release it is on, how far behind main that is, whether
each copy is what this repository made at that commit, and whether the `@fontsource` packages it
installs are the ones `versions.yaml` pinned there. It is a report, and exits 0 whatever it
finds: `task check` does not run it, because a consumer that is behind is no fault of a change
here, and a pull request to this repository cannot fix it.

## Releases

A release is a name for a commit of main that consumers pin, `v<YYYYMMDD>.<NN>` as every
spin-stack repository names its own: the day in UTC and which release of that day. A person cuts
one, after a change consumers should take:

```
go tool task release
```

It refuses a commit `Gate` has not passed on (go-tools' `proven release`) and one already
released, and the GitHub release makes the tag. A consumer that tracks releases pins one in its
`versions.yaml`, and `versions check` and `versions bump brand` move it to the next:

```yaml
- name: brand
  kind: git
  source: https://github.com/spin-stack/brand
  version: v20261122.01
  pin: <the commit that release names>
  track: github-release
```

The `screens-v...` releases are the screens' images and no release of the palette: they are made
with `--latest=false`, so the one GitHub calls latest, which `track: github-release` reads, is
always a `v<YYYYMMDD>.<NN>`.

## License

Apache License 2.0, as spin. The fonts are Inter and JetBrains Mono, each under the SIL Open Font
License (`fonts/LICENSE-Inter.txt`, `fonts/LICENSE-JetBrains-Mono.txt`); `versions.yaml` says
which release of each the files are.
