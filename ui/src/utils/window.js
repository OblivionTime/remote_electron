const roleArg = process.argv.find((arg) => arg.startsWith("--window-role="));

export const windowRole = roleArg ? roleArg.split("=")[1] : "main";
export const isMainWindow = windowRole === "main";
