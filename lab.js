// The laboratory's screens, drawn once for every page that shows them: index.html, the five-second
// test and the recognition test. Every word, row and name here is a fixture: no installation's
// data, no person's name, and nothing that names the product, since the recognition test shows
// these with the mark and the wordmark taken out.

export const esc = (s) => String(s).replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]);
const glyph = (d) => `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${d}</svg>`;
const more = `<button class="btn ghost icon" aria-label="More actions">${glyph('<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>')}</button>`;

// A row's actions: its key and its glyph. A third party's logo would follow its owner's brand
// terms, so the editor is a generic glyph.
const acts = {
	Agent: ['A', '<rect x="4" y="9" width="16" height="11" rx="2"/><path d="M12 9V5"/><circle cx="12" cy="4" r="1"/><path d="M9 14v1M15 14v1"/>'],
	Terminal: ['T', '<path d="m5 8 5 4-5 4"/><path d="M13 17h6"/>'],
	'Open in editor': ['E', '<path d="m16 18 6-6-6-6"/><path d="m8 6-6 6 6 6"/>'],
	Stop: ['S', '<rect x="6" y="6" width="12" height="12" rx="1"/>'],
	Start: ['S', '<path d="M7 4v16l13-8z"/>'],
};
const workspaces = [
	['api', 'ok', 'running', 'acme/api', 'Fixing the flaky login test · working', '8.0 GB', '2 min ago'],
	['agent-1', 'warn', 'waiting', 'acme/api', 'Waiting for a host', '4.0 GB', '9 min ago'],
	['ci-job', 'err', 'failed', 'acme/builder', 'Build stopped at step 4 · needs you', '32.0 GB', '1 h ago'],
	['web', 'ok', 'running', 'acme/website', 'Idle', '4.0 GB', '3 h ago'],
	['docs', 'busy', 'starting', 'acme/website', '—', '2.0 GB', '5 h ago'],
	['notes', 'idle', 'stopped', '—', '—', '0.5 GB', 'yesterday'],
];

// The workspaces as a list. hover, selected and focus are a row's index, and caption the action
// the pointer is on in the hovered row: a state drawn still, for a page that shows each one.
export function list({ hover, selected, focus, caption } = {}) {
	const state = (i) => [i == hover && 'is-hover', i == selected && 'is-selected', i == focus && 'is-focus'].filter(Boolean).join(' ');
	return '<table class="grid"><thead><tr><th>State</th><th>Name</th><th>Project</th><th>Agent</th><th class="n">Memory</th><th class="n">Updated</th></tr></thead><tbody>' +
		workspaces.map(([name, kind, word, project, agent, memory, age], i) =>
			`<tr class="${state(i)}"><td><span class="badge ${kind}">${word}</span></td><td class="name"><a href="#${name}">${name}</a></td><td class="id">${project}</td><td class="muted">${agent}</td><td class="n">${memory}</td><td class="end"><span class="age">${age}</span><span class="acts">` +
			['Agent', 'Terminal', 'Open in editor', kind === 'idle' || kind === 'err' ? 'Start' : 'Stop'].map((a) =>
				`<span class="act${i == hover && a === caption ? ' is-hover' : ''}"><button class="btn sm ghost icon" aria-label="${a}" aria-keyshortcuts="${acts[a][0]}">${glyph(acts[a][1])}</button><span class="cap" aria-hidden="true">${a} <kbd>${acts[a][0]}</kbd></span></span>`).join('') +
			'</span></td></tr>').join('') + '</tbody></table>';
}

// A terminal's lines: [class, text] runs, the classes lab.css gives the ANSI colours.
const line = (...runs) => runs.map((r) => (Array.isArray(r) ? `<span class="${r[0]}">${esc(r[1])}</span>` : esc(r))).join('');
export const session = [
	line(['p', 'dev@api'], ' ', ['b', '~/src/api'], ' $ go test ./login/...'),
	line(['g', 'ok  '], '\tacme/api/login\t0.412s'),
	line(['r', 'FAIL'], '\tacme/api/login/session\t1.208s'),
	line('    session_test.go:88: token refreshed twice in ', ['y', '1.2s']),
	line(['p', 'dev@api'], ' ', ['b', '~/src/api'], ' $ git diff --stat'),
	line(' login/session.go | 14 ', ['g', '+++++++'], ['r', '-----']),
	line(['p', 'dev@api'], ' ', ['b', '~/src/api'], ' $ ') + '<span class="cur"> </span>',
].join('\n');
export const terminal = () => `<pre class="term" role="img" aria-label="A terminal session">${session}</pre>`;

const button = (kind, word, size = '') => `<button class="btn ${size} ${kind}">${word}</button>`;
// fills is how many of a screen's actions are filled. One is the rule ([controls]); three is what
// the five-second test compares it with, the same words in the same places.
const fill = (fills, nth, rest = 'secondary') => (nth <= fills ? 'primary' : rest);

const settingsRow = (label, value, verb) => `<div class="row"><span class="lbl">${label}</span><span class="val">${value}</span>${verb}</div>`;

// What a screen's page holds, by name. Each is { current, main }: the destination the navigation
// marks, and the page.
export const screens = {
	list: ({ fills = 1 } = {}) => ({
		current: 'Workspaces',
		main: `<div class="head"><h1>Workspaces</h1><span class="ctx">6 in 3 projects</span><span class="grow"></span>${button(fill(fills, 3), 'Import')}${button(fill(fills, 2), 'New project')}${button(fill(fills, 1), 'New workspace')}</div>
			<div class="tools"><input placeholder="Search workspaces" aria-label="Search workspaces"></div>
			<div class="grid-wrap">${list()}</div>`,
	}),
	workspace: () => ({
		current: 'Workspaces',
		main: `<div class="head"><h1 class="id">api</h1><span class="badge ok">running</span><span class="ctx"><b>acme/api</b> · on <b>runner-a1</b> · 4 CPU · 8.0 GB</span><span class="grow"></span>${button('secondary', 'Stop')}${button('primary', 'Open terminal')}${more}</div>
			<div class="tabs"><a class="current" href="#">Overview</a><a href="#">Terminal</a><a href="#">Snapshots</a><a href="#">Network</a><a href="#">Settings</a></div>
			<div class="two">
				<div><h3>Agent</h3>${terminal()}</div>
				<div><h3>Details</h3><dl class="facts">
					<div><dt>Image</dt><dd class="id">ghcr.io/acme/sandbox:v12</dd></div>
					<div><dt>Started</dt><dd>Today at <span class="id">14:02:11</span></dd></div>
					<div><dt>Agent</dt><dd>Fixing the flaky login test · working</dd></div>
					<div><dt>Snapshot</dt><dd><span class="id">clean</span> · 2 min ago</dd></div>
					<div><dt>Ports</dt><dd class="id">8080, 5432</dd></div>
				</dl></div>
			</div>
			<h3 class="over">Snapshots</h3>
			<div class="grid-wrap"><table class="grid"><thead><tr><th>Name</th><th>Taken</th><th>Holds</th><th class="n">Size</th></tr></thead><tbody>
				<tr><td class="name"><a href="#clean">clean</a></td><td class="muted">2 min ago</td><td class="muted">Disk and memory</td><td class="n">1.9 GB</td></tr>
				<tr><td class="name"><a href="#before-upgrade">before-upgrade</a></td><td class="muted">Yesterday</td><td class="muted">Disk</td><td class="n">1.2 GB</td></tr>
				<tr><td class="name"><a href="#first-boot">first-boot</a></td><td class="muted">3 days ago</td><td class="muted">Disk</td><td class="n">0.8 GB</td></tr>
			</tbody></table></div>`,
	}),
	settings: ({ fills = 1 } = {}) => ({
		current: 'Settings',
		main: `<div class="head"><h1>Settings</h1><span class="ctx">This installation</span></div>
			<div class="sections">
				<div class="split"><div><h3>General</h3><p class="muted">Its name, and who may sign in.</p></div><div class="card">
					${settingsRow('Name', '<span class="id">acme</span>', button(fill(fills, 2, 'ghost'), 'Rename', 'sm'))}
					${settingsRow('Sign-in', 'Members of <span class="id">acme</span> on the forge', button('ghost', 'Change', 'sm'))}
				</div></div>
				<div class="split"><div><h3>Workspaces</h3><p class="muted">What a new workspace gets, and how long an idle one runs.</p></div><div class="card">
					<div class="row editing"><span class="lbl">Default size</span><span class="val">4 CPU · 8 GB</span><span class="acts">${button('secondary', 'Cancel', 'sm')}${button(fill(fills, 1), 'Save', 'sm')}</span></div>
					${settingsRow('Image', '<span class="id">ghcr.io/acme/sandbox:v12</span>', button('ghost', 'Change', 'sm'))}
					${settingsRow('Stop when idle', 'After 30 minutes', button('ghost', 'Change', 'sm'))}
				</div></div>
				<div class="split"><div><h3>Network</h3><p class="muted">What a workspace may reach. A runner applies a change at once.</p></div><div class="card">
					${settingsRow('Policies', '<span class="id">github</span> <span class="muted">· github.com, ghcr.io</span>', button('ghost', 'Change', 'sm'))}
					${settingsRow('Log connections', '<span class="muted">Off</span>', button(fill(fills, 3, 'ghost'), 'Turn on', 'sm'))}
				</div></div>
			</div>`,
	}),
};

const nav = [
	['Work', [['Workspaces', '<rect x="3" y="4" width="18" height="12" rx="2"/><path d="M8 20h8M12 16v4"/>'], ['Projects', '<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>'], ['Snapshots', '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>']]],
	['Administer', [['Hosts', '<rect x="3" y="4" width="18" height="7" rx="1"/><rect x="3" y="13" width="18" height="7" rx="1"/><path d="M7 7.5h.01M7 16.5h.01"/>'], ['Teams', '<circle cx="9" cy="8" r="3"/><path d="M3 20a6 6 0 0 1 12 0"/><path d="M16 5a3 3 0 0 1 0 6M21 20a6 6 0 0 0-4-5.6"/>'], ['Settings', '<path d="M4 6h10M18 6h2M4 12h4M12 12h8M4 18h12M20 18h0"/><circle cx="16" cy="6" r="2"/><circle cx="10" cy="12" r="2"/><circle cx="18" cy="18" r="2"/>']]],
];
// logos is where the mark's cuts are from the page, and brand whether the frame says whose it is:
// with it off, a block stands where the mark was and another where the wordmark was.
export function app({ current, main }, { brand = true, logos = 'logos/' } = {}) {
	const who = brand
		? `<img class="on-dark" src="${logos}mark-dark-256.png" alt=""><img class="on-light" src="${logos}mark-clear-256.png" alt=""><span>Spinbox</span>`
		: '<i class="no-mark"></i><i class="no-word"></i>';
	return `<div class="app"><nav aria-label="Sample navigation"><div class="who">${who}</div>` +
		nav.map(([group, items]) => `<div class="nav-group">${group}</div>` + items.map(([name, d]) => `<a${name === current ? ' class="current"' : ''} href="#">${glyph(d)}${name}</a>`).join('')).join('') +
		`<div class="foot"><i></i>dev@example.com</div></nav><main><div class="column">${main}</div></main></div>`;
}

// dist/tokens.css gives light's roles to the root alone, and a page that shows light beside dark -
// a terminal on white, light's tints while the page is dark - needs them on an element: the same
// declarations, read from the stylesheet, under .spin-light. Dark has .spin-dark already.
export function scopeLight() {
	for (const sheet of document.styleSheets) {
		if (!sheet.href?.endsWith('tokens.css')) continue;
		for (const rule of sheet.cssRules) {
			if (rule.selectorText !== ":root[data-theme='light']" && rule.selectorText !== ':root[data-theme="light"]') continue;
			const style = document.createElement('style');
			style.textContent = `.spin-light { ${rule.style.cssText} color: var(--spin-text); }\n.spin-dark { color: var(--spin-text); }`;
			document.head.append(style);
			return;
		}
	}
	throw new Error('dist/tokens.css has no rule for light: scopeLight no longer reads how it is written');
}

// What a test page remembers, in this browser alone. Storage can be refused (a private window, a
// blocked site): the page then counts in memory and says so.
export function store(key, initial) {
	let persists = true, value = initial;
	try {
		const raw = localStorage.getItem(key);
		if (raw) value = JSON.parse(raw);
	} catch {
		persists = false;
	}
	return {
		get value() { return value; },
		get persists() { return persists; },
		save(next) {
			value = next;
			try { localStorage.setItem(key, JSON.stringify(value)); } catch { persists = false; }
		},
		clear() {
			value = initial;
			try { localStorage.removeItem(key); } catch { persists = false; }
		},
	};
}
