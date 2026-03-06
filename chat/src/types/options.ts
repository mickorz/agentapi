// 选项相关类型
export interface OptionsItem {
  label: string;
  description?: string;
}

// 选项更新事件
export interface OptionsUpdateEvent {
  message_id: number;
  options: OptionsItem[];
  multi_select: boolean;
  question_id: string;
}
