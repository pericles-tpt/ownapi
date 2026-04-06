import { NodeType } from "./structure";

export function getNodeTypeLabel(type: NodeType): string {
    switch(type) {
        case NodeType.Http:
            return "HTTP"
        case NodeType.Json:
            return "JSON"
        default:
            return "UNKNOWN TYPE"
    }
}
