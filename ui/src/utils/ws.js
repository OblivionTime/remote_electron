const PREFIX = "/v1/api/remote/server";

export const wsUrl = (path, params) => {
  const query = params ? `?${new URLSearchParams(params)}` : "";
  return `ws://${import.meta.env.VITE_API_URL}${PREFIX}${path}${query}`;
};

export const createSocket = (path, params) => new WebSocket(wsUrl(path, params));
