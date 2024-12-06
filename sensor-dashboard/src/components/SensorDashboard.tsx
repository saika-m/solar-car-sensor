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

interface WebSocketMessage {
  data: SensorData;
  connected: boolean;
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

// Component for angle displays (Euler angles)
const AngleGroup = ({ 
  heading = 0, 
  roll = 0, 
  pitch = 0 
}: { 
  heading?: number; 
  roll?: number; 
  pitch?: number; 
}) => (
  <div className="p-4 bg-white rounded-lg shadow-sm">
    <h3 className="text-lg font-semibold mb-2">Euler Angles</h3>
    <div className="grid grid-cols-1 gap-4">
      <div>
        <span className="text-gray-600">Heading:</span>
        <span className="ml-2 font-mono">{formatValue(heading)}°</span>
      </div>
      <div>
        <span className="text-gray-600">Roll:</span>
        <span className="ml-2 font-mono">{formatValue(roll)}°</span>
      </div>
      <div>
        <span className="text-gray-600">Pitch:</span>
        <span className="ml-2 font-mono">{formatValue(pitch)}°</span>
      </div>
    </div>
  </div>
);

// Component for quaternion values
const QuaternionGroup = ({ 
  w = 0, 
  x = 0, 
  y = 0, 
  z = 0 
}: { 
  w?: number; 
  x?: number; 
  y?: number; 
  z?: number; 
}) => (
  <div className="p-4 bg-white rounded-lg shadow-sm">
    <h3 className="text-lg font-semibold mb-2">Quaternion</h3>
    <div className="grid grid-cols-2 gap-4">
      <div>
        <span className="text-gray-600">W:</span>
        <span className="ml-2 font-mono">{formatValue(w, 4)}</span>
      </div>
      <div>
        <span className="text-gray-600">X:</span>
        <span className="ml-2 font-mono">{formatValue(x, 4)}</span>
      </div>
      <div>
        <span className="text-gray-600">Y:</span>
        <span className="ml-2 font-mono">{formatValue(y, 4)}</span>
      </div>
      <div>
        <span className="text-gray-600">Z:</span>
        <span className="ml-2 font-mono">{formatValue(z, 4)}</span>
      </div>
    </div>
  </div>
);

export default function SensorDashboard() {
  const [data, setData] = useState<Partial<SensorData>>({});
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [wsConnected, setWsConnected] = useState(false);

  const connect = useCallback(() => {
    try {
      const ws = new WebSocket('ws://localhost:8080/sensor');

      ws.onopen = () => {
        console.log('WebSocket connected');
        setWsConnected(true);
        setError(null);
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WebSocketMessage;
          setData(message.data);
          setConnected(message.connected);
        } catch (e) {
          console.error('Error parsing data:', e);
          setError('Invalid data received');
        }
      };

      ws.onerror = (event) => {
        console.error('WebSocket error:', event);
        setError('Connection error');
        setWsConnected(false);
        setConnected(false);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        setWsConnected(false);
        setConnected(false);
        // Try to reconnect after 2 seconds
        setTimeout(connect, 2000);
      };

      return ws;
    } catch (e) {
      console.error('WebSocket creation error:', e);
      setError('Failed to create connection');
      setWsConnected(false);
      setConnected(false);
      return null;
    }
  }, []);

  useEffect(() => {
    const ws = connect();
    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, [connect]);

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Solar Car IMU Sensor Dashboard</h1>
      
      <div className="mb-4 flex items-center gap-4">
        <div className="flex items-center">
          <div className={`h-2 w-2 rounded-full ${wsConnected ? 'bg-blue-500' : 'bg-red-500'} mr-2`}></div>
          <span className={`text-sm ${wsConnected ? 'text-blue-700' : 'text-red-700'}`}>
            WebSocket {wsConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>

        <div className="flex items-center">
          <div className={`h-2 w-2 rounded-full ${connected ? 'bg-green-500' : 'bg-yellow-500'} mr-2`}></div>
          <span className={`text-sm ${connected ? 'text-green-700' : 'text-yellow-700'}`}>
            Sensor {connected ? 'Active' : 'No Data'}
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

        <AngleGroup
          heading={data.eulHeading}
          roll={data.eulRoll}
          pitch={data.eulPitch}
        />

        <QuaternionGroup
          w={data.quaW}
          x={data.quaX}
          y={data.quaY}
          z={data.quaZ}
        />
      </div>
    </div>
  );
}