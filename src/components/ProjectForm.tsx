import React, { useState } from 'react';
import './ProjectForm.css';
import { config } from '../config';

// Use only the valid Nobl9 project roles
const ROLE_OPTIONS = [
  { label: 'Owner', value: 'project-owner' },
  { label: 'Editor', value: 'project-editor' },
  { label: 'Viewer', value: 'project-viewer' },
];

interface UserGroup {
  userIds: string; // comma-separated
  role: string;    // backend value
}

const MAX_USERS = 8;

const ProjectForm: React.FC = () => {
  const [appID, setAppID] = useState('');
  const [description, setDescription] = useState('');
  const [userGroups, setUserGroups] = useState<UserGroup[]>([
    { userIds: '', role: ROLE_OPTIONS[0].value },
  ]);
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [message, setMessage] = useState('');
  const [showSummary, setShowSummary] = useState(false);

  // Helper to count total users
  const totalUsers = userGroups.reduce((acc, group) => {
    const ids = group.userIds.split(',').map(id => id.trim()).filter(Boolean);
    return acc + ids.length;
  }, 0);

  // Handlers for user group changes
  const handleUserGroupChange = (idx: number, field: keyof UserGroup, value: string) => {
    const updated = [...userGroups];
    updated[idx][field] = value;
    setUserGroups(updated);
  };

  const handleAddGroup = () => {
    setUserGroups([...userGroups, { userIds: '', role: ROLE_OPTIONS[0].value }]);
  };

  const handleRemoveGroup = (idx: number) => {
    setUserGroups(userGroups.filter((_, i) => i !== idx));
  };

  // Validation
  const validate = (): string | null => {
    if (!appID.trim()) return 'App ID is required.';
    if (!/^[a-z0-9-]+$/.test(appID)) return 'App ID can only contain lowercase letters, numbers, and hyphens.';
    if (totalUsers === 0) return 'At least one user must be specified.';
    if (totalUsers > MAX_USERS) return `Maximum ${MAX_USERS} users allowed per project.`;
    for (const group of userGroups) {
      if (!group.userIds.trim()) return 'User IDs cannot be empty.';
      if (!group.role) return 'Role must be selected for each group.';
    }
    return null;
  };

  // Show summary before submit
  const handleShowSummary = (e: React.FormEvent) => {
    e.preventDefault();
    const error = validate();
    if (error) {
      setStatus('error');
      setMessage(error);
      return;
    }
    setShowSummary(true);
    setStatus('idle');
    setMessage('');
  };

  // Final submit (to be hooked to backend later)
  const handleSubmit = async () => {
    setStatus('loading');
    setMessage('');
    try {
      const response = await fetch('/api/create-project', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ appID, description, userGroups }),
      });
      const data = await response.json();
      if (data.success) {
        setStatus('success');
        setMessage(data.message);
        // Only close modal and reset after a short delay
        setTimeout(() => {
          setShowSummary(false);
          setAppID('');
          setUserGroups([{ userIds: '', role: ROLE_OPTIONS[0].value }]);
          setStatus('idle');
          setMessage('');
        }, 1800);
      } else {
        setStatus('error');
        setMessage(data.message || 'Unknown error');
      }
    } catch (err) {
      setStatus('error');
      setMessage('Failed to connect to backend.');
    }
  };

  return (
    <div className="project-form">
      <form onSubmit={handleShowSummary}>
        <div className="form-group">
          <label htmlFor="appID">App ID:</label>
          <input
            type="text"
            id="appID"
            value={appID}
            onChange={e => setAppID(e.target.value)}
            required
            pattern="[a-z0-9-]+"
            title="Only lowercase letters, numbers, and hyphens allowed"
          />
        </div>
        <div className="form-group">
          <label htmlFor="description">Description:</label>
          <textarea
            id="description"
            value={description}
            onChange={e => setDescription(e.target.value)}
            rows={3}
            placeholder="Enter a description for the project (optional)"
          />
        </div>
        <div className="form-group">
          <label>Enter your email address:</label>
          {userGroups.map((group, idx) => (
            <div key={idx} className="user-group-row">
              <input
                type="text"
                value={group.userIds}
                onChange={e => handleUserGroupChange(idx, 'userIds', e.target.value)}
                required
              />
              <select
                value={group.role}
                onChange={e => handleUserGroupChange(idx, 'role', e.target.value)}
                required
              >
                {ROLE_OPTIONS.map(role => (
                  <option key={role.value} value={role.value}>{role.label}</option>
                ))}
              </select>
              {userGroups.length > 1 && (
                <button type="button" onClick={() => handleRemoveGroup(idx)} className="remove-btn">Remove</button>
              )}
            </div>
          ))}
          <button type="button" onClick={handleAddGroup} disabled={totalUsers >= MAX_USERS} className="add-btn">
            Add another user
          </button>
          <div className="user-count">Total users: {totalUsers} / {MAX_USERS}</div>
        </div>
        {message && <div className={`message ${status}`}>{message}</div>}
        <button type="submit" className="submit-btn">Review & Submit</button>
      </form>
      {showSummary && (
        <div className="summary-modal">
          <div className="summary-card">
            <h3>Confirm Project Details</h3>
            <div className="summary-section">
              <span className="summary-label">App ID:</span>
              <span className="summary-value">{appID}</span>
            </div>
            <div className="summary-section">
              <span className="summary-label">Description:</span>
              <span className="summary-value">{description || <em>No description</em>}</span>
            </div>
            <div className="summary-section">
              <span className="summary-label">Users:</span>
              <ul className="summary-users">
                {userGroups.map((group, idx) => (
                  <li key={idx}>
                    <span className="summary-user-list">{group.userIds}</span>
                    <span className="summary-role">Role: {ROLE_OPTIONS.find(r => r.value === group.role)?.label || group.role}</span>
                  </li>
                ))}
              </ul>
            </div>
            {message && (
              <div className={`message ${status}`}>{message}</div>
            )}
            <div className="summary-actions">
              <button onClick={handleSubmit} className="confirm-btn" disabled={status === 'loading'}>
                {status === 'loading' ? 'Submitting...' : 'Confirm & Submit'}
              </button>
              <button onClick={() => setShowSummary(false)} className="cancel-btn" disabled={status === 'loading'}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
      <div style={{ textAlign: 'center', marginTop: '1.5rem' }}>
        <a
          href={config.helpUrl}
          target="_blank"
          rel="noopener noreferrer"
          style={{ fontSize: '0.95rem', color: '#555', textDecoration: 'underline', opacity: 0.85 }}
        >
          Need help?
        </a>
      </div>
    </div>
  );
};

export default ProjectForm; 