import { useEffect, useState } from 'react';
import { authTokenKey, request } from './api';
import { DataQualityPage, RecommendationQualityPage, SearchAnalytics } from './features/analytics/AnalyticsPages';
import { AuditManagement } from './features/audit/AuditManagement';
import { LoginPage } from './features/auth/LoginPage';
import { EvidenceManagement } from './features/evidence/EvidenceManagement';
import { ImportManagement } from './features/import/ImportManagement';
import { SpeciesManagement } from './features/species/SpeciesManagement';
import { FunctionTagManagement } from './features/tags/FunctionTagManagement';
import { UserManagement } from './features/users/UserManagement';
import type { ActiveMenu, AuthUser } from './types';

const menus: ActiveMenu[] = ['菌种管理', '功能标签', '批量导入', '文献证据', '数据审核', '数据质量', '搜索分析', '账号管理', '推荐质量'];

function menusFor(user: AuthUser): ActiveMenu[] {
  if (user.role === 'admin') return menus;
  if (user.role === 'expert') return ['数据审核'];
  return menus.filter((menu) => !['数据审核', '账号管理'].includes(menu));
}

export function App() {
  const [activeMenu, setActiveMenu] = useState<ActiveMenu>('菌种管理');
  const [user, setUser] = useState<AuthUser | null>(null);
  const [checkingAuth, setCheckingAuth] = useState(true);

  useEffect(() => {
    if (!localStorage.getItem(authTokenKey)) { setCheckingAuth(false); return; }
    void request<AuthUser>('/api/auth/me').then(setUser).catch(() => setUser(null)).finally(() => setCheckingAuth(false));
  }, []);
  useEffect(() => {
    const expire = () => setUser(null);
    window.addEventListener('auth-expired', expire);
    return () => window.removeEventListener('auth-expired', expire);
  }, []);

  if (checkingAuth) return <main className="loginPage"><div className="loginCard">正在验证登录状态…</div></main>;
  if (!user) return <LoginPage onLogin={setUser} />;

  const visibleMenus = menusFor(user);
  const currentMenu = visibleMenus.includes(activeMenu) ? activeMenu : visibleMenus[0];
  function logout() { localStorage.removeItem(authTokenKey); setUser(null); }

  return <main className="layout">
    <aside className="sidebar">
      <h1>运营端</h1>
      <div className="currentUser"><strong>{user.displayName}</strong><span>{user.email} · {user.role}</span></div>
      {visibleMenus.map((menu) => <button className={menu === currentMenu ? 'active' : ''} key={menu} onClick={() => setActiveMenu(menu)}>{menu}</button>)}
      <button className="logoutButton" onClick={logout}>退出登录</button>
    </aside>
    {currentMenu === '菌种管理' && <SpeciesManagement />}
    {currentMenu === '功能标签' && <FunctionTagManagement />}
    {currentMenu === '批量导入' && <ImportManagement />}
    {currentMenu === '文献证据' && <EvidenceManagement />}
    {currentMenu === '数据审核' && <AuditManagement />}
    {currentMenu === '数据质量' && <DataQualityPage />}
    {currentMenu === '搜索分析' && <SearchAnalytics />}
    {currentMenu === '账号管理' && <UserManagement />}
    {currentMenu === '推荐质量' && <RecommendationQualityPage />}
  </main>;
}
