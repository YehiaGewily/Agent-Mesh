import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { AgentType, TaskStatus, type Task } from '../types';
import { clsx } from 'clsx';
import { Activity, CheckCircle, Clock, Eye, X } from 'lucide-react';

interface TaskCardProps {
    task: Task;
}

const item = {
    hidden: { opacity: 0, y: 20 },
    show: { opacity: 1, y: 0 }
};

export const TaskCard: React.FC<TaskCardProps> = ({ task }) => {
    const [showModal, setShowModal] = useState(false);

    const getAgentColor = (type: string) => {
        switch (type) {
            case AgentType.Magnus: return 'border-magnus shadow-glow-magnus text-magnus';
            case AgentType.Cedric: return 'border-cedric shadow-glow-cedric text-cedric';
            case AgentType.Lyra: return 'border-lyra shadow-glow-lyra text-lyra';
            default: return 'border-gray-500 text-gray-400';
        }
    };

    const getAgentBg = (type: string) => {
        switch (type) {
            case AgentType.Magnus: return 'bg-magnus/10';
            case AgentType.Cedric: return 'bg-cedric/10';
            case AgentType.Lyra: return 'bg-lyra/10';
            default: return 'bg-gray-800';
        }
    }

    const getAgentName = (type: string) => {
        switch (type) {
            case AgentType.Magnus: return 'MAGNUS';
            case AgentType.Cedric: return 'CEDRIC';
            case AgentType.Lyra: return 'LYRA';
            default: return 'UNKNOWN';
        }
    }

    const getIcon = () => {
        switch (task.status) {
            case TaskStatus.Pending: return <Clock size={16} />;
            case TaskStatus.Processing: return <Activity size={16} className="animate-pulse" />;
            case TaskStatus.Completed: return <CheckCircle size={16} className="text-green-500" />;
            default: return <Clock size={16} />;
        }
    };

    const calculateDuration = () => {
        if (!task.created_at || !task.updated_at) return null;
        const start = new Date(task.created_at).getTime();
        const end = new Date(task.updated_at).getTime();
        const durationMs = end - start;

        if (durationMs < 1000) return `${durationMs}ms`;
        if (durationMs < 60000) return `${(durationMs / 1000).toFixed(1)}s`;
        return `${(durationMs / 60000).toFixed(1)}m`;
    };

    const isCompleted = task.status === TaskStatus.Completed;

    return (
        <>
            <motion.div
                layoutId={task.id}
                variants={item}
                initial="hidden"
                animate="show"
                exit={{ opacity: 0, scale: 0.9 }}
                className={clsx(
                    "rounded-lg border-l-4 p-4 mb-3 bg-card border border-gray-800 hover:border-r-4 transition-all duration-300 backdrop-blur-sm",
                    getAgentColor(task.agent_type)
                )}
            >
                <div className="flex justify-between items-start mb-2">
                    <span className={clsx("text-xs font-mono px-2 py-0.5 rounded uppercase font-bold tracking-wider", getAgentBg(task.agent_type))}>
                        {getAgentName(task.agent_type)}
                    </span>
                    <div className="flex items-center gap-2">
                        {isCompleted && (
                            <span className="text-xs font-mono px-2 py-0.5 rounded bg-green-500/20 text-green-400 font-bold">
                                SUCCESS
                            </span>
                        )}
                        <span className="text-xs text-gray-500 font-mono">
                            P{task.priority}
                        </span>
                    </div>
                </div>

                <div className="flex items-center gap-2 mb-2">
                    {getIcon()}
                    <span className="text-sm font-semibold truncate text-white">{task.id.substring(0, 8)}...</span>
                </div>

                <div className="text-xs text-gray-400 font-mono break-all mb-2">
                    {JSON.stringify(task.payload).substring(0, 50)}
                    {JSON.stringify(task.payload).length > 50 && "..."}
                </div>

                {isCompleted && (
                    <div className="flex items-center justify-between pt-2 border-t border-gray-800">
                        <span className="text-xs text-gray-500 font-mono">
                            ⏱️ {calculateDuration()}
                        </span>
                        <button
                            onClick={() => setShowModal(true)}
                            className="flex items-center gap-1 text-xs text-magnus hover:text-white transition-colors px-2 py-1 rounded hover:bg-white/5"
                        >
                            <Eye size={12} />
                            Details
                        </button>
                    </div>
                )}
            </motion.div>

            {/* Modal */}
            <AnimatePresence>
                {showModal && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-6"
                        onClick={() => setShowModal(false)}
                    >
                        <motion.div
                            initial={{ scale: 0.9, opacity: 0 }}
                            animate={{ scale: 1, opacity: 1 }}
                            exit={{ scale: 0.9, opacity: 0 }}
                            className="bg-card border border-gray-700 rounded-xl p-6 max-w-2xl w-full max-h-[80vh] overflow-auto"
                            onClick={(e) => e.stopPropagation()}
                        >
                            <div className="flex items-center justify-between mb-4">
                                <h3 className="text-lg font-bold text-white">Mission Details</h3>
                                <button
                                    onClick={() => setShowModal(false)}
                                    className="text-gray-400 hover:text-white transition-colors"
                                >
                                    <X size={20} />
                                </button>
                            </div>

                            <div className="space-y-4">
                                <div>
                                    <label className="text-xs text-gray-500 uppercase tracking-wider">Task ID</label>
                                    <p className="text-sm font-mono text-white mt-1">{task.id}</p>
                                </div>

                                <div>
                                    <label className="text-xs text-gray-500 uppercase tracking-wider">Agent Type</label>
                                    <p className="text-sm font-mono text-white mt-1">{getAgentName(task.agent_type)}</p>
                                </div>

                                <div>
                                    <label className="text-xs text-gray-500 uppercase tracking-wider">Status</label>
                                    <p className="text-sm font-mono text-green-400 mt-1">{task.status}</p>
                                </div>

                                <div>
                                    <label className="text-xs text-gray-500 uppercase tracking-wider">Duration</label>
                                    <p className="text-sm font-mono text-white mt-1">{calculateDuration()}</p>
                                </div>

                                <div>
                                    <label className="text-xs text-gray-500 uppercase tracking-wider">Payload</label>
                                    <pre className="text-xs font-mono text-white mt-2 bg-black/50 p-4 rounded border border-gray-800 overflow-auto">
                                        {JSON.stringify(task.payload, null, 2)}
                                    </pre>
                                </div>

                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="text-xs text-gray-500 uppercase tracking-wider">Created</label>
                                        <p className="text-xs font-mono text-white mt-1">
                                            {new Date(task.created_at).toLocaleString()}
                                        </p>
                                    </div>
                                    <div>
                                        <label className="text-xs text-gray-500 uppercase tracking-wider">Updated</label>
                                        <p className="text-xs font-mono text-white mt-1">
                                            {new Date(task.updated_at).toLocaleString()}
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </>
    );
};
