import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
const { requestMock } = vi.hoisted(()=>({requestMock:vi.fn()}));
vi.mock('../../api',()=>({request:requestMock}));
import { SearchConfigManagement } from './SearchConfigManagement';

describe('SearchConfigManagement',()=>{
  beforeEach(()=>{requestMock.mockReset();requestMock.mockImplementation((path:string,options?:RequestInit)=>{if(path.endsWith('/synonyms')&&!options)return Promise.resolve({items:[{id:'s1',term:'生防',synonym:'病害防控',weight:.85,enabled:true}]});if(path.endsWith('/rules')&&!options)return Promise.resolve({items:[{id:'r1',name:'安全促生',queryPattern:'促生',functionTagCode:'plant-growth-promotion',safetyLevel:'BSL-1',boost:.2,enabled:true}]});if(path.endsWith('/reindex'))return Promise.resolve({indexed:2});return Promise.resolve({})})});
  it('维护同义词、召回规则并重建索引',async()=>{render(<SearchConfigManagement/>);await screen.findByText('生防 → 病害防控');fireEvent.change(screen.getByLabelText('标准词'),{target:{value:'降解'}});fireEvent.change(screen.getByLabelText('同义表达'),{target:{value:'污染去除'}});fireEvent.click(screen.getAllByRole('button',{name:'新增'})[0]);await waitFor(()=>expect(requestMock).toHaveBeenCalledWith('/api/admin/search-config/synonyms',expect.objectContaining({method:'POST'})));fireEvent.change(screen.getByLabelText('规则名称'),{target:{value:'环保召回'}});fireEvent.change(screen.getByLabelText('查询模式'),{target:{value:'废水'}});fireEvent.click(screen.getAllByRole('button',{name:'新增'})[1]);await waitFor(()=>expect(requestMock).toHaveBeenCalledWith('/api/admin/search-config/rules',expect.objectContaining({method:'POST'})));fireEvent.click(screen.getByRole('button',{name:'重建向量索引'}));expect(await screen.findByText('向量索引已更新：2 个菌种')).toBeInTheDocument()});
});
