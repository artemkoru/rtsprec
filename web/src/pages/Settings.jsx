import React, { useState, useEffect, useCallback } from 'react';
import { apiFetch, postJSON } from '../api.js';

// Values the backend expects for the settings the form does not expose yet.
// They match the defaults documented in the README.
const CAMERA_DEFAULTS = {
  motion_sensitivity: 50,
  motion_timeout: 120,
  video_duration_minutes: 15,
  screenshot_interval_minutes: 0,
};

const EMPTY_CAMERA_FORM = { name: '', rtspUrl: '', username: '', password: '' };

export default function Settings() {
  const [cameras, setCameras] = useState([]);
  const [activeTab, setActiveTab] = useState('cameras');
  const [cameraForm, setCameraForm] = useState(EMPTY_CAMERA_FORM);
  const [fieldErrors, setFieldErrors] = useState({});
  const [submitError, setSubmitError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const loadCameras = useCallback(async () => {
    try {
      const data = await apiFetch('/api/cameras');
      setCameras(data || []);
    } catch (error) {
      console.error('Failed to fetch cameras:', error);
    }
  }, []);

  useEffect(() => {
    loadCameras();
  }, [loadCameras]);

  const updateField = (field) => (event) => {
    setCameraForm((form) => ({ ...form, [field]: event.target.value }));
  };

  const validateCameraForm = () => {
    const errors = {};
    if (!cameraForm.name.trim()) errors.name = 'Name is required';
    if (!cameraForm.rtspUrl.trim()) errors.rtspUrl = 'RTSP URL is required';
    return errors;
  };

  const handleAddCamera = async (event) => {
    event.preventDefault();

    const errors = validateCameraForm();
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) {
      return; // nothing is sent until the required fields are filled
    }

    setSubmitting(true);
    setSubmitError('');
    try {
      await postJSON('/api/cameras', {
        name: cameraForm.name.trim(),
        rtsp_url: cameraForm.rtspUrl.trim(),
        username: cameraForm.username,
        password: cameraForm.password,
        ...CAMERA_DEFAULTS,
      });
      setCameraForm(EMPTY_CAMERA_FORM);
      await loadCameras();
    } catch (error) {
      setSubmitError(error.message || 'Could not add the camera');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="container">
      <h1>Settings</h1>

      <div className="tabs">
        {['cameras', 'motion', 'storage', 'scheduling'].map((tab) => (
          <button
            key={tab}
            className={`tab ${activeTab === tab ? 'active' : ''}`}
            onClick={() => setActiveTab(tab)}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      {activeTab === 'cameras' && (
        <div className="tab-content">
          <h3>Camera Configuration</h3>
          <form className="camera-form" onSubmit={handleAddCamera} noValidate>
            <input
              type="text"
              placeholder="Camera Name"
              value={cameraForm.name}
              onChange={updateField('name')}
              aria-invalid={Boolean(fieldErrors.name)}
            />
            {fieldErrors.name && <span className="field-error">{fieldErrors.name}</span>}

            <input
              type="text"
              placeholder="RTSP URL"
              value={cameraForm.rtspUrl}
              onChange={updateField('rtspUrl')}
              aria-invalid={Boolean(fieldErrors.rtspUrl)}
            />
            {fieldErrors.rtspUrl && <span className="field-error">{fieldErrors.rtspUrl}</span>}

            <input
              type="text"
              placeholder="Username (optional)"
              value={cameraForm.username}
              onChange={updateField('username')}
            />
            <input
              type="password"
              placeholder="Password (optional)"
              value={cameraForm.password}
              onChange={updateField('password')}
            />

            <button type="submit" disabled={submitting}>
              {submitting ? 'Adding...' : 'Add Camera'}
            </button>
            {submitError && <p role="alert" className="field-error">{submitError}</p>}
          </form>

          <div className="camera-list">
            {cameras.map((camera) => (
              <div key={camera.id} className="camera-config">
                <h4>{camera.name}</h4>
                <p>Status: {camera.status}</p>
                <button>Edit</button>
                <button>Delete</button>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'motion' && (
        <div className="tab-content">
          <h3>Motion Detection</h3>
          <div>
            <label>Sensitivity (0-100):</label>
            <input type="range" min="0" max="100" defaultValue="50" />
            <span>50</span>
          </div>
          <div>
            <label>Motion Timeout (seconds):</label>
            <input type="number" defaultValue="120" />
          </div>
          <button>Save</button>
        </div>
      )}

      {activeTab === 'storage' && (
        <div className="tab-content">
          <h3>S3 Storage</h3>
          <div>
            <label>S3 Bucket:</label>
            <input type="text" placeholder="my-bucket" />
          </div>
          <div>
            <label>Region:</label>
            <input type="text" placeholder="us-east-1" />
          </div>
          <div>
            <label>Access Key:</label>
            <input type="password" placeholder="AWS Access Key" />
          </div>
          <div>
            <label>Secret Key:</label>
            <input type="password" placeholder="AWS Secret Key" />
          </div>
          <button>Test Connection</button>
          <button>Save</button>
        </div>
      )}

      {activeTab === 'scheduling' && (
        <div className="tab-content">
          <h3>Screenshot Scheduling</h3>
          <div>
            <label>Screenshot Interval (minutes):</label>
            <input type="number" defaultValue="30" />
          </div>
          <button>Save</button>
        </div>
      )}
    </div>
  );
}
