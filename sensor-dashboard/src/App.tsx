import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface SensorData {
  timestamp: number;
  heading: number;
  roll: number;
  pitch: number;
  temperature: number;
  pressure: number;
  altitude: number;
  accX: number;
  accY: number;
  accZ: number;
}

const SensorDashboard: React.FC = () => {
  const [data, setData] = useState<SensorData[]>([]);
  const [latestData, setLatestData] = useState<SensorData | null>(null);

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8080/ws');

    socket.onmessage = (event) => {
      const newData: SensorData = JSON.parse(event.data);
      setLatestData(newData);
      setData((prevData) => [...prevData.slice(-50), newData]);
    };

    return () => {
      socket.close();
    };
  }, []);

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Sensor Dashboard</h1>
      <div className="grid grid-cols-3 gap-4 mb-4">
        <div className="border p-2 rounded">
          <h2 className="font-semibold">Orientation</h2>
          <p>Heading: {latestData?.heading.toFixed(2)}°</p>
          <p>Roll: {latestData?.roll.toFixed(2)}°</p>
          <p>Pitch: {latestData?.pitch.toFixed(2)}°</p>
        </div>
        <div className="border p-2 rounded">
          <h2 className="font-semibold">Environment</h2>
          <p>Temperature: {latestData?.temperature.toFixed(2)}°C</p>
          <p>Pressure: {latestData?.pressure.toFixed(2)} Pa</p>
          <p>Altitude: {latestData?.altitude.toFixed(2)} m</p>
        </div>
        <div className="border p-2 rounded">
          <h2 className="font-semibold">Acceleration</h2>
          <p>X: {latestData?.accX.toFixed(2)} m/s²</p>
          <p>Y: {latestData?.accY.toFixed(2)} m/s²</p>
          <p>Z: {latestData?.accZ.toFixed(2)} m/s²</p>
        </div>
      </div>
      <div className="h-64">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="timestamp" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="heading" stroke="#8884d8" />
            <Line type="monotone" dataKey="roll" stroke="#82ca9d" />
            <Line type="monotone" dataKey="pitch" stroke="#ffc658" />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

export default SensorDashboard;