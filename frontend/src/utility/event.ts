/**
 * 從 Wails Runtime 事件中取出實際 payload（data）
 *
 * ```ts
 * {"name":"image-file-dropped","data":["/Users/ios/Desktop/images.jpg"]}
 * ```
 */
export function eventData<T>(event: unknown): T {

    if (event !== null && typeof event === "object" && "data" in event) {
        return (event as { data: T }).data;
    }

    return event as T;
}