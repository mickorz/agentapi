"use client";

import {useChat} from "./chat-provider";
import MessageInput from "./message-input";
import MessageList from "./message-list";

export function Chat() {
  const {messages, loading, sendMessage, serverStatus, sendAnswer} = useChat();

  return (
    <>
      <MessageList
        messages={messages}
        onAnswer={sendAnswer}
        loading={loading}
      />
      <MessageInput
        onSendMessage={sendMessage}
        disabled={loading}
        serverStatus={serverStatus}
      />
    </>
  );
}
