import React from 'react';
import type { Task } from '../types';
import { TaskCard } from './TaskCard';
import { AnimatePresence, motion } from 'framer-motion';

interface TaskColumnProps {
    title: string;
    tasks: Task[];
    count: number;
    borderColor?: string;
}

export const TaskColumn: React.FC<TaskColumnProps> = ({ title, tasks, count, borderColor = "border-l-gray-800" }) => {
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
                <motion.div layout className="space-y-3">
                    <AnimatePresence mode='popLayout'>
                        {tasks.map((task) => (
                            <TaskCard key={task.id} task={task} />
                        ))}
                    </AnimatePresence>

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
