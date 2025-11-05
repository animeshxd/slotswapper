import type z from "zod";

export type TreeifyError<T> = ReturnType<typeof z.treeifyError<T>>;
