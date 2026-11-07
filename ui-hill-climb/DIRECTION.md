# Direction — round 1

The identity moves from "everything is navy and cyan" to "a neutral, quiet ground with the mark's
blue used rarely". The mark keeps its own colours; the interface stops repeating them everywhere.

## Colour roles

- **Surfaces neutral, faintly cool.** Hue ~215, saturation ≤ 15% (was 39%). Dark:
  page #0E1013 → surface #171A20 → raised #21252C → overlay #2A2F37. Light: page #FFFFFF →
  surface #F0F2F5 → raised #E0E4EA. Each step holds ΔE00 ≥ 3 (the `edge` floor), so planes read as
  planes without a border having to say so.
- **Borders quiet.** `border` is a hairline one step above the surface it is on (#373E48 dark,
  #C3CAD3 light); `border_strong` (an input's edge, 3:1) is a neutral grey (#6A7280 / #7A828E),
  not a blue-grey.
- **One accent, the mark's cyan deepened, as a fill with white text** (#1B7DB4 dark, #0A6FA3
  light). Round 1 tried the mark's royal blue: two of three reviewers read it as generic SaaS and
  the cyan as what made it Spinbox. Links are a lighter/darker tint of the same cyan (#6CC6EE /
  #075A86). The accent is the primary action, the focus ring, the current tab: scarce by rule.
- **Labels are a desaturated slate** (#8FB3C6 / #33586E), not mint: a label names, it does not
  compete with state.
- **States softened, still apart** (ΔE00 ≥ 20 drawn, ≥ 10 in every deficiency): success #4CC38A,
  warning #E3B04B, error #EE7BB0 (magenta-leaning so deutans tell it from success), destructive
  fill #C4384F. Shape still carries state (dot, diamond, square).
- **Contrast is held, not traded**: text 7:1 on every surface, secondary 4.5:1, input edge and
  focus 3:1, in both modes.

## Geometry

- Controls 6 px (`radius_sm`), cards and panels 8 px (`radius`), dialogs and hero figures 12 px
  (`radius_lg`). The dashboard's single `--radius` becomes 6 px: still one radius for a dense
  console, no longer a hard box.

## Type, spacing, density, motion

Unchanged in this round: IBM Plex Sans/Mono, the dense scale, the measure, targets and motion
tokens. The change is colour and corners; if the council says hierarchy still suffers, type is
the next lever.

## Light and dark

Both first-class and derived from the same roles. Dark is a neutral near-black (not navy), light a
white page with neutral greys (not blue-greys).

## Order of propagation

brand (`palette.toml` → `dist/`, the laboratory) → website (Starlight copy) → the app (spin's
brand.css copy). This branch changes brand only.
