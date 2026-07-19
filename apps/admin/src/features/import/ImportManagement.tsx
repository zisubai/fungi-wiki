import { FormEvent, useEffect, useState } from 'react';
import { request } from '../../api';
import { Messages, PageHeader } from '../../components/Common';
import type { ImportBatch, ListResponse } from '../../types';

export function ImportManagement() {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState<ImportBatch | null>(null);
  const [history, setHistory] = useState<ImportBatch[]>([]);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  async function loadHistory() {
    try {
      const data = await request<ListResponse<ImportBatch>>('/api/admin/imports?limit=20');
      setHistory(data.items);
    } catch (e) { setError(e instanceof Error ? e.message : '导入记录加载失败'); }
  }

  useEffect(() => { void loadHistory(); }, []);

  async function upload(event: FormEvent) {
    event.preventDefault();
    if (!file) { setError('请选择 CSV 或 XLSX 文件'); return; }
    setUploading(true); setError(''); setNotice('');
    const body = new FormData(); body.append('file', file);
    try {
      const data = await request<ImportBatch>('/api/admin/imports/species', { method: 'POST', body });
      setResult(data);
      setNotice(`导入完成：成功 ${data.successRows} 行，失败 ${data.failedRows} 行。成功数据已进入待审核队列。`);
      await loadHistory();
    } catch (e) { setError(e instanceof Error ? e.message : '导入失败'); }
    finally { setUploading(false); }
  }

  function downloadTemplate() {
    const headers = ['slug','latin_name','chinese_name','aliases','strain_number','source_environment','safety_level','is_model_organism','summary','function_tags','medium_name','temperature_min','temperature_max','ph_min','ph_max','oxygen_requirement','culture_time'];
    const example = ['bacillus-demo','Bacillus demo','示例芽孢杆菌','旧名称；常用名','CGMCC 1.1','土壤','BSL-1','否','示例数据','biocontrol;plant-growth-promotion','LB','25','37','6.0','8.0','好氧','24 h'];
    const url = URL.createObjectURL(new Blob([`\uFEFF${headers.join(',')}\n${example.join(',')}\n`], { type: 'text/csv;charset=utf-8' }));
    const link = document.createElement('a'); link.href = url; link.download = 'species-import-template.csv'; link.click(); URL.revokeObjectURL(url);
  }

  return <section className="content">
    <PageHeader title="批量导入" description="上传 CSV 或 Excel，逐行校验后进入数据审核流程。" action={<button className="primary" onClick={downloadTemplate}>下载 CSV 模板</button>} />
    <Messages error={error} notice={notice} />
    <section className="importGrid"><form className="panel uploadPanel" onSubmit={upload}><h3>上传菌种数据</h3><p>支持 .csv、.xlsx、.xlsm，最大 10MB、5000 行。功能标签填写编码或名称，多个标签用分号分隔。</p><label className="filePicker"><input type="file" accept=".csv,.xlsx,.xlsm" onChange={(e) => setFile(e.target.files?.[0] ?? null)} /><span>{file?.name ?? '选择文件'}</span></label><button className="primary full" disabled={uploading}>{uploading ? '解析导入中…' : '开始导入并提交审核'}</button></form>
      <div className="panel"><div className="panelTitle"><h3>最近导入</h3><span>{history.length} 个批次</span></div><div className="batchList">{history.map((batch) => <article key={batch.id}><div><strong>{batch.sourceFilename}</strong><small>{new Date(batch.createdAt).toLocaleString()}</small></div><span className={batch.failedRows ? 'countWarning' : 'countSuccess'}>{batch.successRows} 成功 / {batch.failedRows} 失败</span></article>)}{!history.length && <div className="empty">暂无导入记录。</div>}</div></div>
    </section>
    {result?.rows && <section className="panel importResult"><div className="panelTitle"><h3>本次逐行结果</h3><span>{result.totalRows} 行</span></div><div className="tableWrap"><table><thead><tr><th>行号</th><th>Slug</th><th>结果</th><th>问题</th></tr></thead><tbody>{result.rows.map((row) => <tr key={row.rowNumber}><td>{row.rowNumber}</td><td><code>{row.slug || '-'}</code></td><td>{row.status === 'imported' ? '已导入待审核' : '失败'}</td><td className="rowErrors">{row.errors?.join('；') || '-'}</td></tr>)}</tbody></table></div></section>}
  </section>;
}
