---
title: Playground
layout: wide
toc: false
---

<link rel="stylesheet" href="playground.css">
<div id="pg" class="pg">
<p class="pg-intro">Try every operation and preset in your browser. sortmancy runs locally as WebAssembly, so your images never leave your computer. Large images are scaled down for speed; the command below gives you the full-size result.</p>
<div class="pg-grid">
<aside class="pg-panel">
<div class="pg-tabs"><button type="button" data-mode="op" class="pg-active">Operation</button><button type="button" data-mode="pipeline">Pipeline</button></div>
<div id="pg-op-mode">
<label class="pg-field"><span>Operation</span><select id="pg-op"></select></label>
<p id="pg-op-help" class="pg-help"></p>
<div id="pg-params"></div>
<details class="pg-common"><summary>Part of the image</summary><div id="pg-common"></div></details>
</div>
<div id="pg-pipe-mode" hidden>
<label class="pg-field"><span>Pipeline</span><textarea id="pg-pipeline" rows="4" spellcheck="false"></textarea></label>
<label class="pg-field"><span>Preset</span><select id="pg-preset"><option value="">Choose a preset…</option></select></label>
<p class="pg-help">Examples:</p>
<div class="pg-chips" id="pg-examples"></div>
</div>
<div class="pg-row"><label class="pg-field pg-seed"><span>Seed</span><input id="pg-seed" type="number" value="7"></label><button id="pg-random" type="button">Random</button></div>
<div class="pg-row"><label class="pg-check"><input id="pg-auto" type="checkbox" checked><span>Update automatically</span></label><button id="pg-run" class="pg-primary" type="button">Run</button></div>
</aside>
<section class="pg-view">
<div class="pg-toolbar">
<label class="pg-button">Upload image<input id="pg-file" type="file" accept="image/*" hidden></label>
<button id="pg-sample" type="button">Sample photo</button>
<label class="pg-inline">Preview size <select id="pg-size"><option value="640">640 px</option><option value="1280" selected>1280 px</option><option value="1920">1920 px</option></select></label>
<span class="pg-spacer"></span>
<button id="pg-compare" type="button">Hold to compare</button>
<button id="pg-download" type="button">Download PNG</button>
</div>
<div class="pg-stage" id="pg-stage"><canvas id="pg-canvas" width="1280" height="720"></canvas><img id="pg-original" alt="Original image"><div class="pg-drop">Drop an image here</div></div>
<div class="pg-status" id="pg-status">Loading WebAssembly…</div>
<div class="pg-command"><code id="pg-command"></code><button id="pg-copy" type="button">Copy</button></div>
</section>
</div>
</div>
<script src="playground.js" defer></script>
