# Rounds

## 00 — baseline (7e17029)
Navy surfaces at S 39%, cyan accent with dark text, mint labels, 4 px everywhere.
Council (claude, codex, gemini × 2 passes): median 6.0, weakest responsive 4, noise 0.5. All three
name the saturated navy, the accent on every row and the blue-grey light mode.

## 01 — neutral ground, the mark's blue as accent, 6 px controls
palette.toml: surfaces neutral (S ≤ 15%), borders quiet, accent #356BDE/#2F64D8 with white text,
labels slate, states softened; radius_sm 4 → 6; scrim from the dark page; hero on the clear mark.
All 100 pairs hold.

## 02 — selection and the lab's sample follow the app
Neutral hover/selection is `raised` in both themes (light's selected item was invisible); scrim
0.6 → 0.5; the lab's dashboard sample: neutral selection, sans title, Cancel before Delete; ids
and states do not wrap.
Council round 1 (on 01's capture): median 7.0 (+1.0), win 40%, but noise 3.5 (control drift).
Claude and Gemini took the neutral ground in dark; Codex and Claude in light read the royal blue
as generic SaaS. The `nowrap` on ids clipped the lab's narrow table. Gemini's batch b1 did not parse.

## 03 — the accent back to the mark's cyan, buttons, elevation
Accent is the cyan deepened to carry white (#1B7DB4 dark, #0A6FA3 light), links #6CC6EE/#075A86;
selected tint 0.22 → 0.24 to keep its floor. The owner: the dialog's area showed in dark, and an
icon + word on every row was too much. New `[elevation]` (`--spin-shadow-overlay`,
`--shadow-overlay`): a soft shadow in light, only a crisp edge in dark; dialogs on `popover`
(overlay). The specimen and the lab follow the brand's action rule: one primary per view in words
alone, row actions secondary (raised, hairline) plus `⋯`. nowrap reverted.
Council round 3: pairs won by claude 12/12, gemini 10/12 (2 ties), codex 3/12 (8 for the
navy-and-cyan, for identity). Median 7.0. Remaining brand findings: light secondary fills look
disabled, light code a slab, dark planes too close, failed pink reads friendly.

## 04 — planes apart, light steps lighter
Dark surface/raised/overlay one step further apart (#171A20, #21252C, #2A2F37), border #373E48,
border_strong #6A7280 to hold 3:1 on the new raised; error #EE7BB0 (a red fails deutans against
success, so magenta stays). The dashboard's `muted`/`secondary` are the surface in light.
Council round 4 (claude, codex; gemini out of quota): claude 8-2 (2 ties, the two "losses" with
reasons that describe the candidate), codex 6-5. The owner: navy+cyan is the identity, the
problem was contrast - not to be dropped.

## 05 — B, navy ink; buttons are brand's
The owner chose B: the navy darkened to #0A0E16, steps apart, grey lines. [controls] (28/32/36,
icon, gap) and tint.fill: hover and pressed fills derived and held (120 pairs). Lab: Buttons.

## 06 — type system, states, fields, tables, the mark on dark
Font council (claude, codex, 3 screens × 3 matchups at DPR 1): no family wins (Inter–Plex 3-3,
Inter–Geist 4-2, Plex–Geist 4-1); both name the system - 11 px mono states, non-tabular figures,
a 52 px mono headline, uppercase mono labels, broken inline code. Headings are sans
(type.heading), fluid from 390 to 1280 px (type.fluid, also Starlight's --sl-text-h*), figures
tabular, labels sentence case, mono at the size of the text beside it. A state is a badge on its
own tint (tint.badge, held at 4.5:1 over page, surface and raised; light states darkened, success
toward teal for protans): 144 pairs hold. Lab: States, Fields, Tables. logos/mark-dark-*: the
clear mark with its shadow drawn for the dark page.
