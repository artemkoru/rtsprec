import React, { useState, useEffect } from 'react';
import { apiFetch } from '../api.js';

export default function Archive() {
  const [cameras, setCameras] = useState([]);
  const [selectedCamera, setSelectedCamera] = useState(null);
  const [selectedDate, setSelectedDate] = useState(new Date().toISOString().split('T')[0]);
  const [timeline, setTimeline] = useState([]);
  const [showOnlyEvents, setShowOnlyEvents] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchCameras();
  }, []);

  useEffect(() => {
    if (selectedCamera) {
      fetchTimeline();
    }
  }, [selectedCamera, selectedDate]);

  const fetchCameras = async () => {
    try {
      const data = await apiFetch('/api/cameras');
      setCameras(data || []);
      if (data && data.length > 0) {
        setSelectedCamera(data[0].id);
      }
    } catch (error) {
      console.error('Failed to fetch cameras:', error);
    }
  };

  const fetchTimeline = async () => {
    setLoading(true);
    try {
      const data = await apiFetch(
        `/api/archive?camera_id=${selectedCamera}&date=${selectedDate}`
      );
      setTimeline(data || []);
    } catch (error) {
      console.error('Failed to fetch timeline:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredTimeline = showOnlyEvents
    ? timeline.filter((item) => item.type === 'event')
    : timeline;

  return (
    <div className="container">
      <h1>Archive</h1>

      <div className="archive-controls">
        <div>
          <label>Camera: </label>
          <select
            value={selectedCamera || ''}
            onChange={(e) => setSelectedCamera(e.target.value)}
          >
            {cameras.map((camera) => (
              <option key={camera.id} value={camera.id}>
                {camera.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label>Date: </label>
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
          />
        </div>

        <div>
          <label>
            <input
              type="checkbox"
              checked={showOnlyEvents}
              onChange={(e) => setShowOnlyEvents(e.target.checked)}
            />
            Show Only Motion Events
          </label>
        </div>
      </div>

      {loading ? (
        <div>Loading timeline...</div>
      ) : (
        <div className="timeline">
          <h3>{filteredTimeline.length} items</h3>
          {filteredTimeline.map((item, idx) => (
            <div key={idx} className={`timeline-item ${item.type}`}>
              <div className="timeline-time">
                {new Date(item.start_time).toLocaleTimeString()}
              </div>
              <div className="timeline-content">
                {item.type === 'video' ? (
                  <div>
                    Video: {item.duration_seconds}s - {item.video?.file_name}
                  </div>
                ) : (
                  <div>
                    Motion Event (sensitivity: {item.event?.sensitivity_threshold})
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
