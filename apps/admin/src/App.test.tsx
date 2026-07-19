import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { AuthUser } from './types';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('./api', () => ({ authTokenKey: 'fungi_admin_token', request: requestMock }));
vi.mock('./features/auth/LoginPage', () => ({ LoginPage: () => <div>登录页</div> }));
vi.mock('./features/species/SpeciesManagement', () => ({ SpeciesManagement: () => <div>菌种页面</div> }));
vi.mock('./features/tags/FunctionTagManagement', () => ({ FunctionTagManagement: () => <div>标签页面</div> }));
vi.mock('./features/import/ImportManagement', () => ({ ImportManagement: () => <div>导入页面</div> }));
vi.mock('./features/evidence/EvidenceManagement', () => ({ EvidenceManagement: () => <div>证据页面</div> }));
vi.mock('./features/audit/AuditManagement', () => ({ AuditManagement: () => <div>审核页面</div> }));
vi.mock('./features/analytics/AnalyticsPages', () => ({ SearchAnalytics: () => <div>搜索分析</div>, RecommendationQualityPage: () => <div>推荐质量</div> }));
vi.mock('./features/users/UserManagement', () => ({ UserManagement: () => <div>账号页面</div> }));

import { App } from './App';

function renderAs(role: AuthUser['role']) {
  localStorage.setItem('fungi_admin_token', 'token');
  requestMock.mockResolvedValue({ id: '1', email: `${role}@fungi.local`, displayName: role, role });
  render(<App />);
}

describe('App role menus', () => {
  beforeEach(() => requestMock.mockReset());

  it('管理员可访问全部运营菜单', async () => {
    renderAs('admin');
    await waitFor(() => expect(screen.getByText('菌种页面')).toBeInTheDocument());
    for (const menu of ['菌种管理', '功能标签', '批量导入', '文献证据', '数据审核', '搜索分析', '账号管理', '推荐质量']) expect(screen.getByRole('button', { name: menu })).toBeInTheDocument();
  });

  it('专家仅显示数据审核', async () => {
    renderAs('expert');
    expect(await screen.findByText('审核页面')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '数据审核' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '菌种管理' })).not.toBeInTheDocument();
  });

  it('运营人员不能访问审核和账号管理', async () => {
    renderAs('operator');
    await screen.findByText('菌种页面');
    expect(screen.queryByRole('button', { name: '数据审核' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '账号管理' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '批量导入' })).toBeInTheDocument();
  });
});
