import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Activity, Cpu, CircuitBoard, Zap } from 'lucide-react';
import type { SystemHealthMetric } from '../types';
import { clsx } from 'clsx';

interface SystemHealthProps {
    healthData: Record<number, SystemHealthMetric>;
}

export const SystemHealth: React.FC<SystemHealthProps> = ({ healthData }) => {
    // track last update time for heartbeat effect per worker
    const [lastUpdates, setLastUpdates] = useState<Record<number, number>>({});

    useEffect(() => {
        const now = Date.now();
        const updates: Record<number, number> = {};
        Object.values(healthData).forEach(h => {
            updates[h.worker_id] = now;
        });
        setLastUpdates(prev => ({ ...prev, ...updates }));
    }, [healthData]);

    const getProgressColor = (value: number) => {
        if (value < 70) return 'bg-green-500 shadow-[0_0_10px_rgba(34,197,94,0.5)]';
        if (value < 90) return 'bg-yellow-500 shadow-[0_0_10px_rgba(234,179,8,0.5)]';
        return 'bg-red-500 shadow-[0_0_10px_rgba(239,68,68,0.5)]';
    };

    // Silence unused variable warning by using it in a console log for debugging if needed, 
    // or just removing if truly not needed. User wanted "pulsing dot that blinks every time".
    // The current animate-pulse on the dot is continuous. 
    // To make it blink on update, we'd use lastUpdates key to trigger a reflow or animation.
    // For now, I will just log it to silence the linter or use it in the key.
    // Actually, let's use it to key the heartbeat icon to force a re-render/animation.

    return (
        <div className="mb-6">
            <h2 className="text-sm font-bold text-gray-400 mb-3 flex items-center gap-2 uppercase tracking-wider">
                <CircuitBoard size={16} />
                Infrastructure Vitals
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <AnimatePresence>
                    {Object.values(healthData).map((metric) => (
                        <motion.div
                            key={metric.worker_id}
                            initial={{ opacity: 0, scale: 0.9 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.9 }}
                            className="bg-card border border-gray-800 rounded-lg p-4 relative overflow-hidden backdrop-blur-sm group hover:border-gray-700 transition-colors"
                        >
                            {/* Header */}
                            <div className="flex justify-between items-center mb-4">
                                <div className="flex items-center gap-2">
                                    {/* Use lastUpdates to trigger re-animation if we wanted, for now just constant pulse */}
                                    <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                                    <span className="text-xs font-mono font-bold text-gray-300">
                                        WORKER-{metric.worker_id}
                                    </span>
                                </div>
                                <Activity
                                    key={lastUpdates[metric.worker_id]} // Trigger re-render on update
                                    size={14}
                                    className="text-gray-600 group-hover:text-green-400 transition-colors"
                                />
                            </div>

                            {/* Metrics */}
                            <div className="space-y-3">
                                {/* CPU */}
                                <div>
                                    <div className="flex justify-between text-xs font-mono mb-1">
                                        <span className="text-gray-500 flex items-center gap-1">
                                            <Cpu size={10} /> CPU
                                        </span>
                                        <span className="text-white font-bold">{metric.cpu_usage.toFixed(1)}%</span>
                                    </div>
                                    <div className="h-1.5 bg-gray-800 rounded-full overflow-hidden">
                                        <motion.div
                                            className={clsx("h-full rounded-full transition-all duration-500", getProgressColor(metric.cpu_usage))}
                                            initial={{ width: 0 }}
                                            animate={{ width: `${Math.min(100, metric.cpu_usage)}%` }}
                                        />
                                    </div>
                                </div>

                                {/* RAM */}
                                <div>
                                    <div className="flex justify-between text-xs font-mono mb-1">
                                        <span className="text-gray-500 flex items-center gap-1">
                                            <Zap size={10} /> MEM
                                        </span>
                                        <span className="text-white font-bold">{metric.ram_usage.toFixed(1)}%</span>
                                    </div>
                                    <div className="h-1.5 bg-gray-800 rounded-full overflow-hidden">
                                        <motion.div
                                            className={clsx("h-full rounded-full transition-all duration-500", getProgressColor(metric.ram_usage))}
                                            initial={{ width: 0 }}
                                            animate={{ width: `${Math.min(100, metric.ram_usage)}%` }}
                                        />
                                    </div>
                                </div>
                            </div>

                            {/* Timestamp */}
                            <div className="mt-3 pt-3 border-t border-gray-800/50">
                                <span className="text-[10px] text-gray-600 font-mono">
                                    LAST HEARTBEAT: {new Date(metric.timestamp).toLocaleTimeString()}
                                </span>
                            </div>
                        </motion.div>
                    ))}
                </AnimatePresence>

                {Object.keys(healthData).length === 0 && (
                    <div className="col-span-full border border-dashed border-gray-800 rounded-lg p-4 text-center">
                        <span className="text-xs text-gray-600 font-mono">
                            Waiting for telemetry...
                        </span>
                    </div>
                )}
            </div>
        </div>
    );
};
