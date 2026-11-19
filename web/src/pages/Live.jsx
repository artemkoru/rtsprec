import React, { useState, useEffect } from 'react';
import { apiFetch } from '../api.js';

export default function Live() {
  const [cameras, setCameras] = useState([]);
  const [selectedCamera, setSelectedCamera] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCameras();
  }, []);

  const fetchCameras = async () => {
    try {
      const data = await apiFetch('/api/cameras');
      setCameras(data || []);
      if (data && data.length > 0) {
        setSelectedCamera(data[0].id);
      }
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch cameras:', error);
      setLoading(false);
    }
  };

  if (loading) return <div className="container">Loading cameras...</div>;

  return (
    <div className="container">
      <h1>Live View</h1>

      <div className="live-container">
        {selectedCamera && (
          <video
            id="hls-player"
            controls
            width="100%"
            style={{ maxWidth: '800px', marginBottom: '20px' }}
          >
            <source
              src={`/hls/${selectedCamera}/playlist.m3u8`}
              type="application/x-mpegURL"
            />
            Your browser does not support the video tag.
          </video>
        )}

        <div className="camera-list">
          <h3>Cameras</h3>
          {cameras.map((camera) => (
            <div
              key={camera.id}
              className={`camera-item ${selectedCamera === camera.id ? 'active' : ''}`}
              onClick={() => setSelectedCamera(camera.id)}
            >
              <span>{camera.name}</span>
              <span className={`status ${camera.status}`}>{camera.status}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
