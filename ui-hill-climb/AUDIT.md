# Audit — round 0 (brand at 7e17029)

Scope: the brand system as `palette.toml` → `dist/` draws it: the app specimen
(`ui-hill-climb/app.html`, built only from `dist/dashboard.css` the way spin's dashboard uses it),
and the laboratory's site, dashboard and palette sections. Screens are in `rounds/00/`. The real
app's catalog (spin v20261019.02) shows the same problems; the specimen reproduces them.

The owner's words: "not only is it square, the tones are so strong there is no contrast"; Beam's
dark mode "can be understood"; even boxd's is better.

## Colour and contrast (the main problem)

- **Every surface is saturated navy.** page/surface/raised/overlay are hue 221 at S 36–39%
  (#111827 → #273552). Beam's are S ≈ 6%. A saturated dark ground tints every grey, every border
  and the text on it blue, so nothing reads as neutral and the eye has no rest (app-list-dark,
  app-detail-dark, site-dark, dashboard-dark).
- **The accent is a solid, very bright cyan (#27BFE8, S 81%) with dark text**, on every primary
  action and on each row (Open terminal ×2 + Start + New workspace on app-list-dark: four glowing
  blocks). It is the loudest thing on every screen, louder than the data. Beam's fill is a
  mid blue with white text, used once per view.
- **Borders that should separate shout instead**: `--input` is `border_strong` #6A7EA0, a
  blue-grey that every secondary button and every input wears (Stop, View logs, Change size,
  Cancel, the search field). Next to the cyan, the row of actions reads as a row of boxes.
- **Mint labels** (#5FE3D6, S 70%) for eyebrows and section labels: a third saturated hue
  competing with cyan and the status colours (site-dark "THE SITE", "ON THIS PAGE", dashboard-dark
  "WORK").
- **Status colours are neon on navy** (#3DDC97, #FFC857, #F57AB4): legible, but at the same
  intensity as the accent, so state does not stand apart from action.
- **Low perceived contrast despite passing ratios**: the surface steps (page → surface → raised)
  are all hue-shifted navy, so cards and code blocks blend into the page (site-dark's `pre`
  blocks, the dialog on dashboard-dark), while the borders are louder than the surfaces they
  bound. The hierarchy of planes is inverted: lines are loud, planes are flat.
- **Light mode**: raised #DEE6F1 and surface #F0F4F9 are blue-tinted greys; the code block on
  app-detail-light reads as a pale-blue slab. The accent #0A6FA3 teal-blue is fine; border_strong
  #6E7D96 again outlines every secondary button heavily.
- The scrim behind a dialog is the navy at 60%: blue-tinted dim.

## Geometry

- **Every corner of the app is 4 px** (`--radius` = `radius_sm`), dialogs included. With heavy
  borders, buttons and inputs read as hard boxes (app-list-dark actions, app-dialog-dark). Beam's
  controls are ≈6 px, its groups ≈8–10 px; boxd's cards 12 px.
- The site uses pill buttons (999 px) beside 4 px code blocks and 12 px frames: three geometries
  on one page (site-dark hero vs. docs).

## Hierarchy and typography

- Page titles are fine in size, but on app-list-dark the cyan buttons out-rank the title and the
  workspace names; the names (the content) are the same weight as the actions' labels.
- Mono uppercase labels in mint with wide tracking are the most saturated text on the site.
- The site's hero h1 in Plex Mono 600 at ~52 px is heavy; it is a decision of the site, kept for
  now, but the colour around it makes it heavier.

## Consistency

- Selected navigation: in the dashboard it is neutral (raised), in the site sidebar a cyan tint
  with cyan text (site-dark "Snapshots and rollback"); in the lab's dashboard sample cyan again.
  Two rules for the same thing.

## Accessibility (measured, `task report`)

- All 100 pairs hold their floors today. The redesign must keep every floor (body text 7:1,
  secondary 4.5:1, input edge and focus ring 3:1, the ΔE00 distances in every view). Lowering
  saturation must not be bought with lower contrast.

## Generic-UI tells

- None of the usual (gradients, glass, shadows everywhere). The problem is the opposite of
  generic: an identity carried by saturating everything, instead of by a few marks on a neutral
  ground.
