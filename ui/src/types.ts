export const TaskStatus = {
    Pending: "pending",
    Processing: "running",
    Completed: "completed",
    Failed: "failed"
} as const;
export type TaskStatus = typeof TaskStatus[keyof typeof TaskStatus];

export const AgentType = {
    Magnus: "MAGNUS_STRATEGIST",
    Cedric: "CEDRIC_WRITER",
    Lyra: "LYRA_AUDITOR"
} as const;
export type AgentType = typeof AgentType[keyof typeof AgentType];

export interface Task {
    id: string;
    status: TaskStatus;
    priority: number;
    agent_type: string;
    payload: Record<string, any>;
    created_at: string;
    updated_at: string;
    worker_id?: string;
    result?: string;
}

export interface WebSocketMessage {
    type: string;
    payload: any;
}

export interface SystemHealthMetric {
    type: "HEALTH_METRIC";
    worker_id: number;
    cpu_usage: number;
    ram_usage: number;
    timestamp: string;
}

export type WebSocketPayload =
    | { type: "HEALTH_UPDATE"; data: SystemHealthMetric; timestamp: string }
    | { type: undefined;[key: string]: any }; // For existing tasks
