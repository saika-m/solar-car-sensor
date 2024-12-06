'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { AlertCircle } from 'lucide-react';

interface SensorData {
  accX: number; accY: number; accZ: number;
  magX: number; magY: number; magZ: number;
  gyrX: number; gyrY: number; gyrZ: number;
  liaX: number; liaY: number; liaZ: number;
  grvX: number; grvY: number; grvZ: number;
  eulHeading: number; eulRoll: number; eulPitch: number;
  quaW: number; quaX: number; quaY: number; quaZ: number;
}

// Safe number formatter with optional precision
const formatValue = (value: number | undefined | null, precision: number = 2) => {
  if (value === undefined || value === null) return '0.00';
  return value.toFixed(precision);
};

// Component to display a sensor group (3 values)
const SensorGroup = ({ 
  title, 
  x = 0, 
  y = 0, 
  z = 0, 
  unit,
  precision = 2
}: { 
  title: string; 
  x?: number; 
  y?: number; 
  z?: number; 
  unit: string;
  precision?: number;
}) => (
  <div className="p-4 bg-white rounded-lg shadow-sm">
    <h3 className="text-lg font-semibold mb-2">{title}</h3>
    <div className="grid grid-cols-3 gap-4">
      <div>
        <span className="text-gray-600">X:</span>
        <span className="ml-2 font-mono">{formatValue(x, precision)}</span>
      </div>
      <div>
        <span className="text-gray-600">Y:</span>
        <span className="ml-2 font-mono">{formatValue(y, precision)}</span>
      </div>
      <div>
        <span className="text-gray-600">Z:</span>
        <span className="ml-2 font-mono">{formatValue(z, precision)}</span>
      </div>
    </div>
    <div className="text-xs text-gray-500 mt-1">Unit: {unit}</div>
  </div>
);

export default function SensorDashboard() {
  const [data, setData] = useState<SensorData>({
    accX: 0, accY: 0, accZ: 0,
    magX: 0, magY: 0, magZ: 0,
    gyrX: 0, gyrY: 0, gyrZ: 0,
    liaX: 0, liaY: 0, liaZ: 0,
    grvX: 0, grvY: 0, grvZ: 0,
    eulHeading: 0, eulRoll: 0, eulPitch: 0,
    quaW: 0, quaX: 0, quaY: 0, quaZ: 0
  });
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<number>(Date.now());

  useEffect(() => {
    let ws: WebSocket | null = null;
    let checkInterval: NodeJS.Timeout;

    const connect = () => {
      try {
        ws = new WebSocket('ws://localhost:8080/sensor');

        ws.onopen = () => {
          console.log('WebSocket connected');
          setConnected(true);
          setError(null);
        };

        ws.onmessage = (event) => {
          try {
            const newData = JSON.parse(event.data);
            setData(newData);
            setLastUpdate(Date.now());
          } catch (e) {
            console.error('Error parsing data:', e);
          }
        };

        ws.onerror = () => {
          setConnected(false);
          setError('Connection error');
        };

        ws.onclose = () => {
          setConnected(false);
          setTimeout(connect, 2000); // Reconnect after 2 seconds
        };

        // Check for stale data every second
        checkInterval = setInterval(() => {
          if (Date.now() - lastUpdate > 3000) {
            setConnected(false);
          }
        }, 1000);

      } catch (e) {
        setError('Failed to create connection');
        setConnected(false);
      }
    };

    connect();

    // Cleanup on unmount
    return () => {
      if (ws) {
        ws.close();
      }
      if (checkInterval) {
        clearInterval(checkInterval);
      }
    };
  }, [lastUpdate]);

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-2xl font-bold mb-6">Solar Car IMU Sensor Dashboard</h1>
        
        <div className="mb-4 flex items-center gap-4">
          <div className="flex items-center">
            <div className={`h-2 w-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'} mr-2`}></div>
            <span className={`text-sm ${connected ? 'text-green-700' : 'text-red-700'}`}>
              {connected ? 'Connected' : 'Disconnected'}
            </span>
          </div>

          {error && (
            <div className="flex items-center text-sm text-red-600">
              <AlertCircle className="h-4 w-4 mr-1" />
              {error}
            </div>
          )}
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <SensorGroup
            title="Accelerometer"
            x={data.accX}
            y={data.accY}
            z={data.accZ}
            unit="mg"
          />
          
          <SensorGroup
            title="Magnetometer"
            x={data.magX}
            y={data.magY}
            z={data.magZ}
            unit="µT"
          />
          
          <SensorGroup
            title="Gyroscope"
            x={data.gyrX}
            y={data.gyrY}
            z={data.gyrZ}
            unit="dps"
          />
          
          <SensorGroup
            title="Linear Acceleration"
            x={data.liaX}
            y={data.liaY}
            z={data.liaZ}
            unit="mg"
          />
          
          <SensorGroup
            title="Gravity Vector"
            x={data.grvX}
            y={data.grvY}
            z={data.grvZ}
            unit="mg"
          />

          <SensorGroup
            title="Euler Angles"
            x={data.eulHeading}
            y={data.eulRoll}
            z={data.eulPitch}
            unit="deg"
          />

          <div className="p-4 bg-white rounded-lg shadow-sm">
            <h3 className="text-lg font-semibold mb-2">Quaternion</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-gray-600">W:</span>
                <span className="ml-2 font-mono">{formatValue(data.quaW, 4)}</span>
              </div>
              <div>
                <span className="text-gray-600">X:</span>
                <span className="ml-2 font-mono">{formatValue(data.quaX, 4)}</span>
              </div>
              <div>
                <span className="text-gray-600">Y:</span>
                <span className="ml-2 font-mono">{formatValue(data.quaY, 4)}</span>
              </div>
              <div>
                <span className="text-gray-600">Z:</span>
                <span className="ml-2 font-mono">{formatValue(data.quaZ, 4)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}