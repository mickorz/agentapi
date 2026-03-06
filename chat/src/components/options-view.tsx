"use client";

import React, { useState } from "react";
import { Button } from "./ui/button";
import { CheckCircle2, Circle } from "lucide-react";

import { useChat } from "./chat-provider";

import { OptionsItem } from "../types/options";

interface OptionsViewProps {
  options: OptionsItem[];
  multiSelect: boolean;
  questionId: string;
  onAnswer: (questionId: string, answers: number[]) => void;
  disabled?: boolean;
}

export function OptionsView({
  options,
  multiSelect,
  questionId,
  onAnswer,
  disabled = false,
}: OptionsViewProps) {
  const [selectedIndices, setSelectedIndices] = useState<Set<number>>(new Set());

  const handleOptionClick = (index: number) => {
    if (disabled) return;

    setSelectedIndices((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(index)) {
        newSet.delete(index);
      } else {
        if (multiSelect) {
          newSet.add(index);
        } else {
          // 单选模式: 清空并只选当前
          newSet.clear();
          newSet.add(index);
        }
      }
      return newSet;
    });
  };

  const handleSubmit = () => {
    if (selectedIndices.size > 0) {
      const answers = Array.from(selectedIndices).sort((a, b) => a - b);
      onAnswer(questionId, answers);
      // 清空选择状态
      setSelectedIndices(new Set());
    }
  };

  return (
    <div className="flex flex-col gap-2 mt-3 mb-4 p-4 border rounded-lg bg-muted/30">
      <div className="flex flex-wrap gap-2">
        {options.map((option, index) => (
          <button
            key={index}
            onClick={() => handleOptionClick(index)}
            disabled={disabled}
            className={`
              flex items-center gap-2 px-4 py-2 rounded-lg border transition-all
              ${
                selectedIndices.has(index)
                  ? "bg-primary text-primary-foreground border-primary"
                  : "bg-background hover:bg-accent border-border"
              }
              ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}
            `}
          >
            {selectedIndices.has(index) ? (
              <CheckCircle2 className="h-4 w-4" />
            ) : (
              <Circle className="h-4 w-4 opacity-30" />
            )}
            <span className="text-sm font-medium">{option.label}</span>
          </button>
        ))}
      </div>
      {options.some((o) => o.description) && selectedIndices.size > 0 && (
        <div className="text-xs text-muted-foreground mt-1">
          {options[Array.from(selectedIndices)[0]]?.description}
        </div>
      )}
      <div className="flex justify-end mt-3">
        <Button
          onClick={handleSubmit}
          disabled={disabled || selectedIndices.size === 0}
          size="sm"
        >
          Submit Selection
        </Button>
      </div>
    </div>
  );
}
