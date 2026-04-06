import { uuid } from "./alias";

// TODO: This should be retrieved from the backend in future
export const ZERO_UUID: uuid = "00000000-0000-0000-0000-000000000000";
export const MAX_RENDERED_CARDS     = 10;
export const GO_ZERO_TIME = new Date(-62135596800000);
export const DEFAULT_PAGE_SIZE = 20;
export const WAIT_FOR_SOCKET_OPEN_SECONDS = 0.25;
export const dummyResponse = new Promise<Response>((res, rej) => res(new Response(null, { status: 200, statusText: ""})))