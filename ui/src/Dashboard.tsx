import React, { useEffect, useState } from 'react';
import { TaskStatus, type Task, type SystemHealthMetric, type WebSocketPayload } from './types';
import { TaskColumn } from './components/TaskColumn';
import { SystemHealth } from './components/SystemHealth';

const Dashboard: React.FC = () => {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [workerHealth, setWorkerHealth] = useState<Record<number, SystemHealthMetric>>({});
    const [isConnected, setIsConnected] = useState(false);

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
                console.log("WebSocket Message:", rawData);

                // Handle Health Updates
                if (rawData.type === "HEALTH_UPDATE") {
                    const healthMsg = rawData as WebSocketPayload & { type: "HEALTH_UPDATE" };
                    setWorkerHealth(prev => ({
                        ...prev,
                        [healthMsg.data.worker_id]: healthMsg.data
                    }));
                    return;
                }

                const data = rawData; // Fallback for existing task objects

                // Handle different message formats
                let updatedTask: Task;

                // Case 1: Full Task object (has 'id')
                if (data.id && data.status && data.agent_type) {
                    updatedTask = data as Task;
                }
                // Case 2: Partial Update/Backend Event (has 'task_id')
                // We map 'task_id' to 'id' to avoid crashes, but we might be missing other fields.
                else if (data.task_id) {
                    updatedTask = {
                        id: data.task_id,
                        status: data.status,
                        // Provide defaults for missing fields to prevent rendering crashes
                        agent_type: data.agent_type || 'UNKNOWN',
                        priority: data.priority || 0,
                        payload: data.payload || {},
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString()
                    } as Task;
                } else {
                    console.warn("Received invalid task format:", data);
                    return;
                }

                setTasks(prevTasks => {
                    const existingTaskIndex = prevTasks.findIndex(t => t.id === updatedTask.id);
                    if (existingTaskIndex >= 0) {
                        const newTasks = [...prevTasks];
                        // Merge existing with update to prevent data loss (e.g. payload)
                        newTasks[existingTaskIndex] = { ...newTasks[existingTaskIndex], ...updatedTask };
                        return newTasks;
                    } else {
                        return [...prevTasks, updatedTask];
                    }
                });

            } catch (err) {
                console.error('Failed to parse WebSocket message:', err);
            }
        };

        return () => {
            ws.close();
        };
    }, []);

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
                    </div>
                </div>
                <div className="flex gap-4 text-xs font-mono text-gray-400">
                    <div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-magnus"></div> Magnus</div>
                    <div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-cedric"></div> Cedric</div>
                    <div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-lyra"></div> Lyra</div>
                </div>
            </header>

            <SystemHealth healthData={workerHealth} />

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 h-[calc(100vh-200px)]">
                <TaskColumn
                    title="PENDING QUEUE"
                    tasks={pendingTasks}
                    count={pendingTasks.length}
                    borderColor="border-l-blue-500/30"
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
                />
            </div>
        </div>
    );
};

export default Dashboard;
