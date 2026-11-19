import React, { useState, useEffect } from 'react';

export default function Settings() {
  const [cameras, setCameras] = useState([]);
  const [activeTab, setActiveTab] = useState('cameras');

  useEffect(() => {
    fetchCameras();
  }, []);

  const fetchCameras = async () => {
    try {
      const response = await fetch('/api/cameras');
      const data = await response.json();
      setCameras(data || []);
    } catch (error) {
      console.error('Failed to fetch cameras:', error);
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
          <div className="camera-form">
            <input type="text" placeholder="Camera Name" />
            <input type="text" placeholder="RTSP URL" />
            <input type="text" placeholder="Username (optional)" />
            <input type="password" placeholder="Password (optional)" />
            <button>Add Camera</button>
          </div>

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
