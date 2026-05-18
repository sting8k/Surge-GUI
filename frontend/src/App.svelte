<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { ListDownloads, AddURL, AddDownload, PauseDownload, ResumeDownload, DeleteDownload, GetBackendStatus, OpenDirectoryDialog, GetStartupEnabled, SetStartupEnabled, GetIconMode, SetIconMode, GetVersion, OpenFile, OpenInFinder, UpdateDownloadURL, PauseAll, ResumeAll, ClearCompleted } from '../wailsjs/go/main/App';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import type { main } from '../wailsjs/go/models';

  // ── State ──
  let downloads: main.DownloadItem[] = [];
  let urlInput = '';
  let serverConnected = false;
  let backendStatus: main.BackendStatus | null = null;
  let backendMessage = 'checking engine...';
  let totalSpeed = 0;
  let selectedId: string | null = null;
  let activeTab = 1; // 0=Queued, 1=Active, 2=Done
  let launchAtLogin = false;
  let startupBusy = false;
  let iconMode: 'dock' | 'menu_bar' = 'dock';
  let iconModeBusy = false;
  let appVersion = 'dev';
  let bulkBusy = false;

  // Modal
  let showAddModal = false;
  let modalURL = '';
  let modalPath = '';
  let modalFilename = '';
  let modalMirrors = '';

  // Recovery modal
  let showUpdateUrlModal = false;
  let updateUrlId = '';
  let updateUrlValue = '';

  // Toasts
  let toasts: Array<{id: number; message: string; type: string}> = [];
  let toastId = 0;

  // Context menu
  let ctxMenu: { x: number; y: number; dl: main.DownloadItem } | null = null;

  function openCtxMenu(e: MouseEvent, dl: main.DownloadItem) {
    e.preventDefault();
    e.stopPropagation();
    ctxMenu = { x: e.clientX, y: e.clientY, dl };
  }

  function closeCtxMenu() {
    ctxMenu = null;
  }

  async function ctxOpen() {
    if (!ctxMenu) return;
    const path = ctxMenu.dl.dest_path;
    closeCtxMenu();
    if (!path) { toast('path unknown', 'error'); return; }
    try { await OpenFile(path); } catch (e: any) { toast(`open failed: ${e}`, 'error'); }
  }

  async function ctxOpenInFinder() {
    if (!ctxMenu) return;
    const path = ctxMenu.dl.dest_path;
    closeCtxMenu();
    if (!path) { toast('path unknown', 'error'); return; }
    try { await OpenInFinder(path); } catch (e: any) { toast(`reveal failed: ${e}`, 'error'); }
  }

  async function ctxCopyURL() {
    if (!ctxMenu) return;
    const url = ctxMenu.dl.url;
    closeCtxMenu();
    try { await navigator.clipboard.writeText(url); toast('url copied', 'info'); } catch { toast('copy failed', 'error'); }
  }

  async function ctxDelete() {
    if (!ctxMenu) return;
    const id = ctxMenu.dl.id;
    closeCtxMenu();
    await doDelete(id);
  }

  // ── Lifecycle ──
  let pollInterval: ReturnType<typeof setInterval>;
  let cleanups: Array<() => void> = [];
  const timeoutHandles = new Set<ReturnType<typeof setTimeout>>();
  let refreshListTimer: ReturnType<typeof setTimeout> | null = null;

  onMount(async () => {
    document.addEventListener('click', closeCtxMenu);
    document.addEventListener('contextmenu', (e) => {
      // Only allow contextmenu on dl-items (handled by openCtxMenu), block everywhere else
      if (!(e.target as HTMLElement)?.closest('.dl-item')) {
        e.preventDefault();
      }
    });
    await Promise.all([
      refreshBackendStatus(),
      refreshStartupState(),
      refreshIconMode(),
      refreshList()
    ]);
    try { appVersion = await GetVersion(); } catch {}

    pollInterval = setInterval(async () => {
      await Promise.all([
        refreshBackendStatus(),
        refreshList()
      ]);
    }, 4000);

    cleanups.push(EventsOn('surge:progress', (data: any) => updateProgress(data)));
    cleanups.push(EventsOn('surge:download_progress', (data: any) => updateProgress(data)));
    cleanups.push(EventsOn('surge:started', (data: any) => {
      toast(`⬇ ${data.Filename || data.filename || 'download'}`, 'info');
      scheduleRefreshList();
    }));
    const onComplete = (data: any) => {
      toast(`✔ ${data.Filename || data.filename || 'download'}`, 'success');
      scheduleRefreshList();
    };
    cleanups.push(EventsOn('surge:complete', onComplete));
    cleanups.push(EventsOn('surge:download_complete', onComplete));
    cleanups.push(EventsOn('surge:error', (data: any) => {
      toast(`✖ ${data.Filename || data.filename}`, 'error');
      scheduleRefreshList();
    }));
    cleanups.push(EventsOn('surge:paused', () => scheduleRefreshList()));
    cleanups.push(EventsOn('surge:resumed', () => scheduleRefreshList()));
    cleanups.push(EventsOn('surge:removed', () => scheduleRefreshList()));
    cleanups.push(EventsOn('surge:queued', () => scheduleRefreshList()));
  });

  onDestroy(() => {
    document.removeEventListener('click', closeCtxMenu);
    if (pollInterval) clearInterval(pollInterval);
    if (refreshListTimer) clearTimeout(refreshListTimer);
    for (const handle of timeoutHandles) {
      clearTimeout(handle);
    }
    timeoutHandles.clear();
    cleanups.forEach(fn => fn());
  });

  // ── Data ──
  async function refreshBackendStatus() {
    try {
      backendStatus = await GetBackendStatus();
      serverConnected = !!backendStatus?.running;
      backendMessage = backendStatus?.message || (serverConnected ? 'Surge backend is ready.' : 'Surge backend is offline.');
    } catch (e: any) {
      backendStatus = null;
      serverConnected = false;
      backendMessage = `Surge backend status failed: ${e}`;
    }
  }

  async function refreshList() {
    try {
      const list = await ListDownloads();
      downloads = list || [];
      totalSpeed = downloads.filter(d => statusKey(d.status) === 'downloading').reduce((s, d) => s + (d.speed || 0), 0);
    } catch { downloads = []; }
  }

  function scheduleRefreshList(delayMs = 0) {
    if (refreshListTimer) {
      clearTimeout(refreshListTimer);
      refreshListTimer = null;
    }

    const handle = setTimeout(async () => {
      timeoutHandles.delete(handle);
      refreshListTimer = null;
      await refreshList();
    }, delayMs);

    timeoutHandles.add(handle);
    refreshListTimer = handle;
  }

  function updateProgress(data: any) {
    const id = data.DownloadID || data.id;
    if (!id) return;
    const speed = data.Speed || data.speed || 0;
    const dl = data.Downloaded || data.downloaded || 0;
    const total = data.Total || data.total || 0;
    downloads = downloads.map(d => d.id === id ? {
      ...d,
      speed: speed / 1048576,
      downloaded: dl,
      total_size: total > 0 ? total : d.total_size,
      progress: total > 0 ? (dl / total) * 100 : d.progress,
      connections: data.ActiveConnections || data.connections || d.connections,
      status: 'downloading'
    } : d);
    totalSpeed = downloads.filter(d => statusKey(d.status) === 'downloading').reduce((s, d) => s + (d.speed || 0), 0);
  }

  // ── Filtering by tab ──
  function statusKey(s: string): string {
    const k = (s || 'queued').toLowerCase().trim();
    if (!k || k === 'pending') return 'queued';
    if (k === 'pausing') return 'paused';
    if (k === 'failed' || k === 'errored') return 'error';
    if (k === 'done' || k === 'complete' || k === 'finished' || k === 'success') return 'completed';
    return k;
  }

  function filterByTab(items: main.DownloadItem[], tab: number): main.DownloadItem[] {
    return items.filter(d => {
      const s = statusKey(d.status);
      if (tab === 0) return s === 'queued';
      if (tab === 1) return s === 'downloading' || s === 'paused' || s === 'error';
      return s === 'completed';
    });
  }

  // ── Actions ──
  async function quickAdd() {
    const url = urlInput.trim();
    if (!url) return;
    if (!serverConnected) { toast(backendMessage || 'engine offline', 'error'); return; }
    try {
      await AddURL(url);
      urlInput = '';
      toast('queued', 'info');
      scheduleRefreshList(500);
    } catch (e: any) {
      if (String(e).includes('409')) toast('already exists', 'info');
      else toast(`failed: ${e}`, 'error');
    }  }

  function onKey(e: KeyboardEvent) { if (e.key === 'Enter') quickAdd(); }

  async function submitAdd() {
    if (!modalURL.trim()) return;
    if (!serverConnected) { toast(backendMessage || 'engine offline', 'error'); return; }
    try {
      await AddDownload(modalURL.trim(), modalPath, modalFilename, modalMirrors);
      showAddModal = false;
      modalURL = modalPath = modalFilename = modalMirrors = '';
      toast('queued', 'info');
      scheduleRefreshList(500);
    } catch (e: any) {
      if (String(e).includes('409')) toast('already exists', 'info');
      else toast(`failed: ${e}`, 'error');
    }  }

  async function browseDir() {
    try { const d = await OpenDirectoryDialog(); if (d) modalPath = d; } catch {}
  }

  async function doPause(id: string) { try { await PauseDownload(id); scheduleRefreshList(); } catch {} }
  async function doResume(id: string) { try { await ResumeDownload(id); scheduleRefreshList(); } catch {} }
  async function doDelete(id: string) {
    try {
      await DeleteDownload(id);
      if (selectedId === id) selectedId = null;
      scheduleRefreshList();
    } catch (e: any) { toast(`delete failed: ${e}`, 'error'); }
  }

  function openUpdateURL(dl: main.DownloadItem) {
    updateUrlId = dl.id;
    updateUrlValue = dl.url || '';
    showUpdateUrlModal = true;
  }

  async function submitUpdateURL() {
    if (!updateUrlId || !updateUrlValue.trim()) return;
    try {
      await UpdateDownloadURL(updateUrlId, updateUrlValue.trim());
      await ResumeDownload(updateUrlId);
      showUpdateUrlModal = false;
      updateUrlId = '';
      updateUrlValue = '';
      toast('url updated and resumed', 'success');
      scheduleRefreshList(500);
    } catch (e: any) {
      toast(`update failed: ${e}`, 'error');
    }
  }

  async function runBulk(label: string, action: () => Promise<main.BulkActionResult>) {
    if (bulkBusy) return;
    bulkBusy = true;
    try {
      const result = await action();
      toast(`${label}: ${result.succeeded}/${result.attempted}`, result.errors?.length ? 'error' : 'success');
      scheduleRefreshList(500);
    } catch (e: any) {
      toast(`${label} failed: ${e}`, 'error');
    } finally {
      bulkBusy = false;
    }
  }

  async function refreshStartupState() {
    try {
      launchAtLogin = await GetStartupEnabled();
    } catch {}
  }

  async function refreshIconMode() {
    try {
      const mode = await GetIconMode();
      iconMode = mode === 'menu_bar' ? 'menu_bar' : 'dock';
    } catch {}
  }

  async function setIconMode(mode: 'dock' | 'menu_bar') {
    if (iconModeBusy || iconMode === mode) return;
    iconModeBusy = true;
    try {
      await SetIconMode(mode);
      await refreshIconMode();
      toast(`icon mode: ${iconMode === 'menu_bar' ? 'menu bar' : 'dock'}`, 'info');
    } catch (e: any) {
      await refreshIconMode();
      toast(`icon mode failed: ${e}`, 'error');
    } finally {
      iconModeBusy = false;
    }
  }

  async function toggleLaunchAtLogin() {
    if (startupBusy) return;
    startupBusy = true;
    const next = !launchAtLogin;
    try {
      await SetStartupEnabled(next);
      await refreshStartupState();
      toast(launchAtLogin ? 'launch at login: on' : 'launch at login: off', 'info');
    } catch (e: any) {
      await refreshStartupState();
      toast(`startup setting failed: ${e}`, 'error');
    } finally {
      startupBusy = false;
    }
  }

  // ── Formatting ──
  function sz(b: number): string {
    if (!b || b <= 0) return '—';
    if (b < 1024) return `${b}B`;
    if (b < 1048576) return `${(b / 1024).toFixed(1)}K`;
    if (b < 1073741824) return `${(b / 1048576).toFixed(1)}M`;
    return `${(b / 1073741824).toFixed(2)}G`;
  }

  function spd(m: number): string {
    if (!m || m <= 0) return '—';
    if (m >= 1024) return `${(m / 1024).toFixed(1)} G/s`;
    if (m >= 1) return `${m.toFixed(1)} M/s`;
    return `${(m * 1024).toFixed(0)} K/s`;
  }

  function fmtETA(seconds: number): string {
    if (!seconds || seconds <= 0) return '—';
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
    return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
  }

  function statusIcon(s: string): string {
    const k = statusKey(s);
    if (k === 'downloading') return '⬇';
    if (k === 'completed') return '✔';
    if (k === 'paused') return '⏸';
    if (k === 'error') return '✖';
    return '⋯';
  }

  function statusColor(s: string): string {
    const k = statusKey(s);
    if (k === 'downloading') return 'var(--st-downloading)';
    if (k === 'completed') return 'var(--st-completed)';
    if (k === 'paused') return 'var(--st-paused)';
    if (k === 'error') return 'var(--st-error)';
    return 'var(--st-queued)';
  }

  function toast(msg: string, type: string) {
    const id = ++toastId;
    toasts = [...toasts, { id, message: msg, type }];
    setTimeout(() => { toasts = toasts.filter(t => t.id !== id); }, 3000);
  }

  // ── Derived ──
  $: activeCount = downloads.filter(d => statusKey(d.status) === 'downloading' || statusKey(d.status) === 'paused' || statusKey(d.status) === 'error').length;
  $: queuedCount = downloads.filter(d => statusKey(d.status) === 'queued').length;
  $: doneCount = downloads.filter(d => statusKey(d.status) === 'completed').length;
  $: filtered = filterByTab(downloads, activeTab);
  $: selected = selectedId ? downloads.find(d => d.id === selectedId) : null;
  $: totalDownloaded = downloads.reduce((s, d) => s + (d.downloaded || 0), 0);
</script>

<div class="titlebar"></div>

<div class="app-root">
  <!-- TOP BAR -->
  <div class="top-bar">
    <div class="brand">
      surge<span class="sep">/</span><span class="sub">downloads</span>
    </div>
    <div class="right">
      {#if totalSpeed > 0}
        <span>↓ <span class="speed">{spd(totalSpeed)}</span></span>
      {/if}
      {#if activeCount > 0}
        <span>{activeCount} active</span>
      {/if}
      <span>
        <span class="dot" class:on={serverConnected} class:off={!serverConnected}></span>
        {serverConnected ? 'engine' : 'offline'}
      </span>
    </div>
  </div>

  <!-- INPUT BAR -->
  <div class="input-bar">
    <div class="input-wrap">
      <span class="prefix">$</span>
      <input type="text" placeholder={serverConnected ? 'paste url' : 'engine offline'} bind:value={urlInput} on:keydown={onKey} id="url-input" disabled={!serverConnected} />
    </div>
    <button class="btn btn-accent" on:click={quickAdd} id="btn-add" disabled={!serverConnected}>Add</button>
    <button class="btn" on:click={() => showAddModal = true} id="btn-advanced" title="Advanced" disabled={!serverConnected}>⋯</button>
  </div>

  {#if !serverConnected}
    <div class="engine-banner">
      <strong>Engine offline.</strong> {backendMessage}
      {#if backendStatus?.cli_version}<span>CLI: {backendStatus.cli_version}</span>{/if}
    </div>
  {/if}

  <!-- DASHBOARD: 2-column -->
  <div class="dashboard">
    <!-- LEFT COLUMN -->
    <div class="col-left">
      <!-- TAB BAR -->
      <div class="tab-bar">
        <button class="tab" class:active={activeTab === 1} on:click={() => activeTab = 1}>
          Active<span class="count">{activeCount}</span>
        </button>
        <button class="tab" class:active={activeTab === 0} on:click={() => activeTab = 0}>
          Queued<span class="count">{queuedCount}</span>
        </button>
        <button class="tab" class:active={activeTab === 2} on:click={() => activeTab = 2}>
          Done<span class="count">{doneCount}</span>
        </button>
      </div>

      <!-- DOWNLOAD LIST -->
      <div class="dl-list">
        {#if filtered.length === 0}
          <div class="dl-list-empty">
            <div class="prompt">$ surge add &lt;url&gt;<span class="cur"></span></div>
            <p>
              {#if activeTab === 0}No queued downloads{:else if activeTab === 2}No completed downloads{:else}No active downloads{/if}
            </p>
          </div>
        {:else}
          {#each filtered as dl (dl.id)}
            {@const sk = statusKey(dl.status)}
            <div
              class="dl-item"
              class:selected={selectedId === dl.id}
              on:click={() => selectedId = dl.id}
              on:contextmenu={(e) => openCtxMenu(e, dl)}
              on:keydown={() => {}}
              role="button"
              tabindex="0"
            >
              <span class="status-icon" style="color: {statusColor(dl.status)}">{statusIcon(dl.status)}</span>
              <div class="info">
                <div class="fname" title={dl.filename}>{dl.filename || 'resolving...'}</div>
                <div class="meta">
                  <span>{dl.progress > 0 ? dl.progress.toFixed(1) + '%' : '—'}</span>
                  <span>·</span>
                  <span>{sz(dl.downloaded)}{#if dl.total_size > 0}/{sz(dl.total_size)}{/if}</span>
                </div>
              </div>
              {#if sk === "completed"}
                <button class="btn-icon" on:click|stopPropagation={() => OpenInFinder(dl.dest_path)} title="Open in Finder"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg></button>
              {:else if sk === "downloading"}
                <span class="spd">{spd(dl.speed)}</span>
              {:else}
                <span class="spd">—</span>
              {/if}
              <div class="acts">
                {#if sk === 'downloading'}
                  <button class="btn-icon" on:click|stopPropagation={() => doPause(dl.id)} title="Pause">⏸</button>
                {:else if sk === 'paused'}
                  <button class="btn-icon" on:click|stopPropagation={() => doResume(dl.id)} title="Resume">▶</button>
                {:else if sk === 'error'}
                  <button class="btn-icon" on:click|stopPropagation={() => openUpdateURL(dl)} title="Update URL and retry">↻</button>
                {/if}
                <button class="btn-icon" on:click|stopPropagation={() => doDelete(dl.id)} title="Delete" style="color: var(--st-error)">✕</button>
              </div>
              <!-- Mini progress -->
              <div class="mini-progress {sk}" style="width: {sk === 'completed' ? 100 : Math.min(dl.progress || 0, 100)}%"></div>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <!-- RIGHT COLUMN -->
    <div class="col-right">
      <!-- FILE DETAILS HEADER -->
      <div class="panel-header">
        <span class="panel-title">File Details</span>
      </div>

      <!-- FILE DETAILS -->
      <div class="detail-panel">
        {#if selected}
          {@const sk = statusKey(selected.status)}
          <!-- Status badge -->
          <div style="margin-bottom: 10px;">
            <span class="status-badge {sk}">
              {statusIcon(selected.status)} {sk}
            </span>
          </div>

          <!-- File info -->
          <div class="detail-section">
            <div class="label">Filename</div>
            <div class="value">{selected.filename || 'resolving...'}</div>
          </div>

          <div class="detail-section">
            <div class="label">URL</div>
            <div class="value url">{selected.url}</div>
          </div>

          {#if selected.dest_path}
            <div class="detail-section">
              <div class="label">Path</div>
              <div class="value">{selected.dest_path}</div>
            </div>
          {/if}

          <div class="detail-section">
            <div class="label">ID</div>
            <div class="value" style="color: var(--text-dim); font-size: 9px;">{selected.id}</div>
          </div>

          <!-- Progress bar -->
          <div class="detail-progress">
            <div class="progress-track">
              <div
                class="progress-fill {sk}"
                style="width: {sk === 'completed' ? 100 : Math.min(selected.progress || 0, 100)}%"
              ></div>
            </div>
          </div>

          <!-- Stats grid -->
          <div class="stats-grid">
            <div class="stat-row">
              <span class="stat-label">Size</span>
              <span class="stat-value">
                {sz(selected.downloaded)}{#if selected.total_size > 0} / {sz(selected.total_size)}{/if}
              </span>
            </div>
            <div class="stat-row">
              <span class="stat-label">Speed</span>
              <span class="stat-value speed">
                {#if sk === 'downloading'}{spd(selected.speed)}{:else if sk === 'completed' && selected.avg_speed > 0}{spd(selected.avg_speed / 1048576)} avg{:else}—{/if}
              </span>
            </div>
            {#if sk === 'downloading' && selected.connections > 0}
              <div class="stat-row">
                <span class="stat-label">Conns</span>
                <span class="stat-value">{selected.connections}</span>
              </div>
            {/if}
            {#if sk === 'downloading' && selected.eta > 0}
              <div class="stat-row">
                <span class="stat-label">ETA</span>
                <span class="stat-value accent">{fmtETA(selected.eta)}</span>
              </div>
            {/if}
            {#if selected.time_taken > 0}
              <div class="stat-row">
                <span class="stat-label">Time</span>
                <span class="stat-value">{fmtETA(selected.time_taken)}</span>
              </div>
            {/if}
          </div>

          {#if selected.error}
            <div class="detail-section" style="margin-top: 10px;">
              <div class="label">Error</div>
              <div class="value" style="color: var(--st-error); font-size: 10px;">{selected.error}</div>
            </div>
          {/if}

          <!-- Action buttons -->
          <div class="detail-actions">
            {#if sk === 'downloading'}
              <button class="btn btn-pause" on:click={() => doPause(selected.id)}>⏸ Pause</button>
            {:else if sk === 'paused'}
              <button class="btn btn-resume" on:click={() => doResume(selected.id)}>▶ Resume</button>
            {:else if sk === 'error'}
              <button class="btn btn-resume" on:click={() => openUpdateURL(selected)}>↻ Update URL</button>
            {/if}
            <button class="btn btn-delete" on:click={() => doDelete(selected.id)}>✕ Delete</button>
          </div>
        {:else}
          <div class="detail-empty">Select a download to view details</div>
        {/if}
      </div>

      <!-- NETWORK STATS PANEL -->
      <div class="stats-panel">
        <div class="panel-header">
          <span class="panel-title">Network</span>
        </div>
        <div class="content">
          <div class="stats-card">
            <span class="label">Speed</span>
            <span class="value cyan">{totalSpeed > 0 ? spd(totalSpeed) : '—'}</span>
          </div>
          <div class="stats-card">
            <span class="label">Active</span>
            <span class="value pink">{activeCount}</span>
          </div>
          <div class="stats-card">
            <span class="label">Downloaded</span>
            <span class="value purple">{sz(totalDownloaded)}</span>
          </div>
        </div>
        <div class="bulk-actions">
          <button class="btn" disabled={bulkBusy || activeCount === 0} on:click={() => runBulk('pause all', PauseAll)}>pause all</button>
          <button class="btn" disabled={bulkBusy || activeCount === 0} on:click={() => runBulk('resume all', ResumeAll)}>resume all</button>
          <button class="btn" disabled={bulkBusy || doneCount === 0} on:click={() => runBulk('clear done', ClearCompleted)}>clear done</button>
        </div>
      </div>
    </div>
  </div>

  <!-- FOOTER -->
  <div class="footer">
    <div class="left">
      <span class:connected={serverConnected} class:disconnected={!serverConnected}>
        {serverConnected ? '● connected' : '○ offline'}
      </span>
      <span>{backendStatus?.base_url || ':1700'}</span>
      {#if backendStatus?.cli_version}<span>{backendStatus.cli_version}</span>{/if}
    </div>
    <div class="right">
      <div class="mode-switch" role="group" aria-label="icon mode">
        <button
          class="footer-toggle"
          aria-pressed={iconMode === 'dock'}
          on:click={() => setIconMode('dock')}
          disabled={iconModeBusy}
        >dock</button>
        <button
          class="footer-toggle"
          aria-pressed={iconMode === 'menu_bar'}
          on:click={() => setIconMode('menu_bar')}
          disabled={iconModeBusy}
        >bar</button>
      </div>
      <button
        class="footer-toggle"
        on:click={toggleLaunchAtLogin}
        disabled={startupBusy}
        aria-pressed={launchAtLogin}
      >
        startup {startupBusy ? '...' : (launchAtLogin ? 'on' : 'off')}
      </button>
      <span>{downloads.length} items</span>
      <span class="version">v{appVersion}</span>
    </div>
  </div>
</div>

<!-- TOASTS -->
<div class="toast-wrap">
  {#each toasts as t (t.id)}
    <div class="toast {t.type}">{t.message}</div>
  {/each}
</div>

<!-- CONTEXT MENU -->
{#if ctxMenu}
  {@const sk = statusKey(ctxMenu.dl.status)}
  <div class="ctx-menu" style="left:{ctxMenu.x}px;top:{ctxMenu.y}px">
    {#if sk === 'completed'}
      <button class="ctx-item" on:click={ctxOpen}>Open</button>
      <button class="ctx-item" on:click={ctxOpenInFinder}>Open in Finder</button>
      <div class="ctx-sep"></div>
      <button class="ctx-item" on:click={ctxCopyURL}>Copy URL</button>
      <div class="ctx-sep"></div>
      <button class="ctx-item danger" on:click={ctxDelete}>Delete</button>
    {:else}
      <button class="ctx-item" on:click={ctxCopyURL}>Copy URL</button>
    {/if}
  </div>
{/if}

<!-- UPDATE URL MODAL -->
{#if showUpdateUrlModal}
  <div class="modal-bg" on:click|self={() => showUpdateUrlModal = false} on:keydown={(e) => e.key === 'Escape' && (showUpdateUrlModal = false)}>
    <div class="modal-box">
      <h2>Update URL</h2>
      <p class="modal-help">Use this for failed or expired download links. The replacement URL must point to the same file.</p>
      <div class="form-group">
        <label class="form-label" for="update-url">new url</label>
        <input class="form-input" id="update-url" type="text" placeholder="https://..." bind:value={updateUrlValue} on:keydown={(e) => e.key === 'Enter' && submitUpdateURL()} />
      </div>
      <div class="modal-actions">
        <button class="btn" on:click={() => showUpdateUrlModal = false}>cancel</button>
        <button class="btn btn-accent" on:click={submitUpdateURL}>update & resume</button>
      </div>
    </div>
  </div>
{/if}

<!-- ADD MODAL -->
{#if showAddModal}
  <div class="modal-bg" on:click|self={() => showAddModal = false} on:keydown={(e) => e.key === 'Escape' && (showAddModal = false)}>
    <div class="modal-box">
      <h2>Add Download</h2>
      <div class="form-group">
        <label class="form-label" for="modal-url">url</label>
        <input class="form-input" id="modal-url" type="text" placeholder="https://..." bind:value={modalURL} on:keydown={(e) => e.key === 'Enter' && submitAdd()} />
      </div>
      <div class="form-group">
        <label class="form-label" for="modal-filename">filename</label>
        <input class="form-input" id="modal-filename" type="text" placeholder="auto" bind:value={modalFilename} />
      </div>
      <div class="form-group">
        <label class="form-label" for="modal-path">save to</label>
        <div style="display:flex;gap:6px">
          <input class="form-input" id="modal-path" type="text" placeholder="default" bind:value={modalPath} />
          <button class="btn" on:click={browseDir}>browse</button>
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="modal-mirrors">mirrors</label>
        <textarea class="form-input form-textarea" id="modal-mirrors" placeholder="one mirror URL per line" bind:value={modalMirrors}></textarea>
      </div>
      <div class="modal-actions">
        <button class="btn" on:click={() => showAddModal = false}>cancel</button>
        <button class="btn btn-accent" on:click={submitAdd}>download</button>
      </div>
    </div>
  </div>
{/if}
