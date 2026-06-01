<template>
  <div class="app">
    <header class="header">
      <div class="container header-inner">
        <router-link to="/" class="logo">
          <span class="logo-icon">📦</span>
          <span>报关 Excel 生成系统</span>
        </router-link>
        <nav>
          <router-link to="/" class="nav-link">上传文件</router-link>
        </nav>
      </div>
    </header>
    <main class="main">
      <router-view />
    </main>
    <footer class="footer">
      <div class="container">
        <p>基于 HS CODE 自动生成报关 Excel · Vue + Go</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
  font-size: 17px;
  color: var(--text);
}

.logo-icon {
  font-size: 22px;
}

.nav-link {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  padding: 6px 12px;
  border-radius: 6px;
}

.nav-link.router-link-active {
  color: var(--primary);
  background: var(--primary-light);
}

.main {
  flex: 1;
  padding: 40px 0 60px;
}

.footer {
  border-top: 1px solid var(--border);
  padding: 20px 0;
  text-align: center;
  color: var(--text-secondary);
  font-size: 13px;
}
</style>
