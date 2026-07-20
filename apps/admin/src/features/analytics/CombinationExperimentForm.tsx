import { FormEvent, useState } from 'react';
import { request } from '../../api';

type Candidate = { members: { latinName: string }[] };
type Props = { recordId: string; candidates: Candidate[]; onSaved: () => void };

export function CombinationExperimentForm({ recordId, candidates, onSaved }: Props) {
  const [candidateIndex, setCandidateIndex] = useState(0);
  const [outcome, setOutcome] = useState('compatible');
  const [temperature, setTemperature] = useState('');
  const [ph, setPH] = useState('');
  const [notes, setNotes] = useState('');
  const [message, setMessage] = useState('');
  const [saving, setSaving] = useState(false);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setMessage('');
    try {
      await request(`/api/admin/recommendations/combinations/${recordId}/experiments`, {
        method: 'POST',
        body: JSON.stringify({
          candidateIndex,
          outcome,
          temperature: temperature === '' ? null : Number(temperature),
          ph: ph === '' ? null : Number(ph),
          notes,
        }),
      });
      setNotes('');
      setMessage('实验结果已记录');
      onSaved();
    } catch (cause) {
      setMessage(cause instanceof Error ? cause.message : '实验结果保存失败');
    } finally {
      setSaving(false);
    }
  }

  return <form className="experimentForm" onSubmit={(event) => void submit(event)}>
    <strong>记录共培养实验</strong>
    <select aria-label="被验证组合" value={candidateIndex} onChange={(event) => setCandidateIndex(Number(event.target.value))}>
      {candidates.map((candidate, index) => <option value={index} key={index}>#{index + 1} {candidate.members.map((member) => member.latinName).join(' + ')}</option>)}
    </select>
    <select aria-label="实验结论" value={outcome} onChange={(event) => setOutcome(event.target.value)}>
      <option value="compatible">兼容</option><option value="incompatible">不兼容</option><option value="inconclusive">结论不明确</option>
    </select>
    <input aria-label="实验温度" type="number" step="0.1" placeholder="温度 °C" value={temperature} onChange={(event) => setTemperature(event.target.value)} />
    <input aria-label="实验 pH" type="number" min="0" max="14" step="0.1" placeholder="pH" value={ph} onChange={(event) => setPH(event.target.value)} />
    <input aria-label="实验备注" maxLength={2000} placeholder="培养基、时长、观察结果等" value={notes} onChange={(event) => setNotes(event.target.value)} />
    <button className="primary" disabled={saving || !candidates.length}>{saving ? '保存中…' : '保存结果'}</button>
    {message && <small>{message}</small>}
  </form>;
}
