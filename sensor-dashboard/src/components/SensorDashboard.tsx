import { useState, useEffect } from 'react';
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

const SensorGroup = ({ title, x, y, z, unit }: { title: string; x: number; y: number; z: number; unit: string }) => (
  <div className="p-4 bg-white rounded-lg shadow-sm">
    <h3 className="text-lg font-semibold mb-2">{title}</h3>
    <div className="grid grid-cols-3 gap-4">
      <div>
        <span className="text-gray-600">X:</span>
        <span className="ml-2 font-mono">{x.toFixed(2)}</span>
      </div>
      <div>
        <span className="text-gray-600">Y:</span>
        <span className="ml-2 font-mono">{y.toFixed(2)}</span>
      </div>
      <div>
        <span className="text-gray-600">Z:</span>
        <span className="ml-2 font-mono">{z.toFixed(2)}</span>
      </div>
    </div>
    <div className="text-xs text-gray-500 mt-1">Unit: {unit}</div>
  </div>
);

const SensorDashboard = () => {
  const [data, setData] = useState<SensorData | null>(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let ws: WebSocket;
    let reconnectTimeout: NodeJS.Timeout;

    const connectWebSocket = () => {
      ws = new WebSocket('ws://localhost:8080/ws');

      ws.onopen = () => {
        setConnected(true);
        setError(null);
        console.log('Connected to sensor server');
      };

      ws.onmessage = (event) => {
        try {
          const newData = JSON.parse(event.data);
          setData(newData);
        } catch (e) {
          console.error('Error parsing sensor data:', e);
        }
      };

      ws.onerror = (event) => {
        setError('Failed to connect to sensor server');
        setConnected(false);
        console.error('WebSocket error:', event);
      };

      ws.onclose = () => {
        setConnected(false);
        console.log('Disconnected from sensor server');
        
        // Try to reconnect after 5 seconds
        reconnectTimeout = setTimeout(connectWebSocket, 5000);
      };
    };

    connectWebSocket();

    // Cleanup on component unmount
    return () => {
      if (ws) {
        ws.close();
      }
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
    };
  }, []);

  if (!connected) {
    return (
      <div className="p-4">
        <div className="flex items-center p-4 bg-red-50 border border-red-200 rounded-lg">
          <AlertCircle className="h-4 w-4 text-red-600 mr-2" />
          <div>
            <h3 className="font-medium text-red-800">Connection Error</h3>
            <p className="text-red-700">
              {error || 'Disconnected from sensor server. Please check if the server is running.'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="p-4">
        <div className="flex items-center p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
          <div>
            <p className="text-blue-700">Loading sensor data...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">IMU Sensor Dashboard</h1>
      <div className="mb-4 flex items-center">
        <div className="h-2 w-2 rounded-full bg-green-500 mr-2"></div>
        <span className="text-sm text-green-700">Connected to sensor</span>
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

        <div className="p-4 bg-white rounded-lg shadow-sm">
          <h3 className="text-lg font-semibold mb-2">Euler Angles</h3>
          <div className="grid grid-cols-1 gap-4">
            <div>
              <span className="text-gray-600">Heading:</span>
              <span className="ml-2 font-mono">{data.eulHeading.toFixed(2)}°</span>
            </div>
            <div>
              <span className="text-gray-600">Roll:</span>
              <span className="ml-2 font-mono">{data.eulRoll.toFixed(2)}°</span>
            </div>
            <div>
              <span className="text-gray-600">Pitch:</span>
              <span className="ml-2 font-mono">{data.eulPitch.toFixed(2)}°</span>
            </div>
          </div>
        </div>

        <div className="p-4 bg-white rounded-lg shadow-sm">
          <h3 className="text-lg font-semibold mb-2">Quaternion</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <span className="text-gray-600">W:</span>
              <span className="ml-2 font-mono">{data.quaW.toFixed(3)}</span>
            </div>
            <div>
              <span className="text-gray-600">X:</span>
              <span className="ml-2 font-mono">{data.quaX.toFixed(3)}</span>
            </div>
            <div>
              <span className="text-gray-600">Y:</span>
              <span className="ml-2 font-mono">{data.quaY.toFixed(3)}</span>
            </div>
            <div>
              <span className="text-gray-600">Z:</span>
              <span className="ml-2 font-mono">{data.quaZ.toFixed(3)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SensorDashboard;