import React from 'react';
import type { Task } from '../types';
import { TaskCard } from './TaskCard';
import { motion } from 'framer-motion';

interface TaskColumnProps {
    title: string;
    tasks: Task[];
    count: number;
    borderColor?: string;
    limit?: number; // Optional limit for "Visual Virtualization"
}

export const TaskColumn: React.FC<TaskColumnProps> = ({ title, tasks, count, borderColor = "border-l-gray-800", limit }) => {
    // Determine how many to show. If limit is undefined, show all (or default slice if we want safety, but let's respect the plan which says "Active: All")
    // Actually, plan says Active: All. So if limit is undefined, we show all.
    // But we previously hardcoded slice(0, 50). Let's use limit || length.
    const showCount = limit || tasks.length;
    const visibleTasks = tasks.slice(0, showCount);
    const hiddenCount = tasks.length - visibleTasks.length;

    return (
        <div className={`bg-card/50 rounded-xl border border-gray-800/50 backdrop-blur-sm flex flex-col h-full overflow-hidden ${borderColor} border-l-4`}>
            {/* Header */}
            <div className="p-4 border-b border-gray-800/50 flex justify-between items-center bg-black/20">
                <div className="flex items-center justify-between">
                    <h2 className="text-sm font-bold uppercase tracking-widest text-gray-300">{title}</h2>
                    <span className="flex items-center justify-center w-6 h-6 text-xs font-bold rounded-full bg-white/10 text-white">
                        {count}
                    </span>
                </div>
            </div>

            {/* List */}
            <div className="flex-1 p-4 overflow-y-auto min-h-0">
                <motion.div className="space-y-3">
                    {visibleTasks.map((task) => (
                        <TaskCard key={task.id} task={task} />
                    ))}

                    {hiddenCount > 0 && (
                        <div className="text-center py-4">
                            <span className="text-xs font-mono text-gray-500 bg-gray-800/50 px-3 py-1 rounded-full border border-gray-700">
                                +{hiddenCount} more in queue
                            </span>
                        </div>
                    )}

                    {tasks.length === 0 && (
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 0.5 }}
                            className="text-center py-10 text-gray-600 text-sm font-mono border-2 border-dashed border-gray-800 rounded-lg"
                        >
                            No Tasks
                        </motion.div>
                    )}
                </motion.div>
            </div>
        </div>
    );
};
