import React, { useEffect, useState } from 'react';
import { TaskStatus, type Task, type SystemHealthMetric, type WebSocketPayload } from './types';
import { TaskColumn } from './components/TaskColumn';
import { SystemHealth } from './components/SystemHealth';

const Dashboard: React.FC = () => {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [workerHealth, setWorkerHealth] = useState<Record<number, SystemHealthMetric>>({});
    const [isConnected, setIsConnected] = useState(false);
    const [highTrafficMode, setHighTrafficMode] = useState(false);

    const messageBuffer = React.useRef<any[]>([]);

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8081/v1/ws');

        ws.onopen = () => {
            console.log('Connected to WebSocket');
            setIsConnected(true);
        };

        ws.onclose = () => {
            console.log('Disconnected from WebSocket');
            setIsConnected(false);
        };

        ws.onmessage = (event) => {
            try {
                const rawData = JSON.parse(event.data);
                messageBuffer.current.push(rawData);
            } catch (err) {
                console.error('Failed to parse WebSocket message:', err);
            }
        };

        return () => {
            ws.close();
        };
    }, []);

    // Flush buffer every 100ms (10 FPS)
    useEffect(() => {
        const interval = setInterval(() => {
            const bufferSize = messageBuffer.current.length;

            // Traffic Detection: if we processed more than 50 messages in this tick
            if (bufferSize > 50) {
                setHighTrafficMode(true);
            } else if (bufferSize === 0 && highTrafficMode) {
                // Optional: cooldown or keep it on for a bit. For now, immediate off when idle.
                // Actually, let's keep it sticky for a moment or just check if buffer is consistently low.
                // Simple approach: off if 0.
                setHighTrafficMode(false);
            }

            if (bufferSize === 0) return;

            const messages = [...messageBuffer.current];
            messageBuffer.current = []; // Clear buffer

            setTasks(prevTasks => {
                let newTasks = [...prevTasks];
                let healthUpdates: Record<number, SystemHealthMetric> = {};
                let hasHealthUpdates = false;

                messages.forEach(data => {
                    // Handle Health Updates
                    if (data.type === "HEALTH_UPDATE") {
                        const healthMsg = data as WebSocketPayload & { type: "HEALTH_UPDATE" };
                        healthUpdates[healthMsg.data.worker_id] = healthMsg.data;
                        hasHealthUpdates = true;
                        return;
                    }

                    // Handle Tasks
                    let updatedTask: Task;
                    if (data.id && data.status && data.agent_type) {
                        updatedTask = data as Task;
                    } else if (data.task_id) {
                        updatedTask = {
                            id: data.task_id,
                            status: data.status,
                            agent_type: data.agent_type || 'UNKNOWN',
                            priority: data.priority || 0,
                            payload: data.payload || {},
                            created_at: new Date().toISOString(),
                            updated_at: new Date().toISOString()
                        } as Task;
                    } else {
                        return;
                    }

                    const existingTaskIndex = newTasks.findIndex(t => t.id === updatedTask.id);
                    if (existingTaskIndex >= 0) {
                        newTasks[existingTaskIndex] = { ...newTasks[existingTaskIndex], ...updatedTask };
                    } else {
                        newTasks.push(updatedTask);
                    }
                });

                if (hasHealthUpdates) {
                    setWorkerHealth(prev => ({ ...prev, ...healthUpdates }));
                }

                return newTasks;
            });
        }, 100);

        return () => clearInterval(interval);
    }, [highTrafficMode]);

    const pendingTasks = tasks.filter(t => t.status === TaskStatus.Pending);
    const activeTasks = tasks.filter(t => t.status === TaskStatus.Processing);
    const completedTasks = tasks.filter(t => t.status === TaskStatus.Completed || t.status === TaskStatus.Failed);

    return (
        <div className="min-h-screen bg-background text-white p-8 font-sans">
            <header className="mb-8 flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent flex items-center gap-2">
                        <span className="bg-blue-500 text-white px-2 py-1 rounded text-sm font-mono">AM</span>
                        Agent Mesh <span className="font-thin text-white">Command Center</span>
                    </h1>
                    <div className="flex items-center gap-2 mt-2 ml-1">
                        <span className="text-xs text-gray-500 uppercase tracking-widest font-mono">System Status:</span>
                        <span className={`text-xs font-bold font-mono ${isConnected ? 'text-green-500 shadow-glow-green' : 'text-red-500'}`}>
                            {isConnected ? 'ONLINE' : 'OFFLINE'}
                        </span>
                        {highTrafficMode && (
                            <span className="ml-4 text-xs font-bold font-mono text-amber-500 animate-pulse border border-amber-500/50 px-2 py-0.5 rounded bg-amber-500/10">
                                ⚠️ HIGH TRAFFIC
                            </span>
                        )}
                    </div>
                </div>
                <div className="flex gap-4 text-xs font-mono text-gray-400">
                    <div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-orange-500"></div> Architect</div>
                    <div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-blue-500"></div> Developer</div>
                    <div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-green-500"></div> QA</div>
                </div>
            </header>

            <SystemHealth healthData={workerHealth} />

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 h-[calc(100vh-200px)]">
                <TaskColumn
                    title="PENDING QUEUE"
                    tasks={pendingTasks}
                    count={pendingTasks.length}
                    borderColor="border-l-blue-500/30"
                    limit={20}
                />
                <TaskColumn
                    title="ACTIVE OPERATIONS"
                    tasks={activeTasks}
                    count={activeTasks.length}
                    borderColor="border-l-green-500/30"
                />
                <TaskColumn
                    title="MISSION HISTORY"
                    tasks={completedTasks}
                    count={completedTasks.length}
                    borderColor="border-l-purple-500/30"
                    limit={15}
                />
            </div>
        </div>
    );
};

export default Dashboard;
