import type { ReactNode } from 'react';

export function PageHeader({ title, description, action }: { title: string; description: string; action?: ReactNode }) {
  return <header className="header"><div><p className="eyebrow">Admin Console</p><h2>{title}</h2><p>{description}</p></div>{action}</header>;
}

export function Messages({ error, notice }: { error: string; notice: string }) {
  return <>{error && <div className="message error">{error}</div>}{notice && <div className="message success">{notice}</div>}</>;
}
