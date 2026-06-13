import { NodeType } from "./structure";

export function getNodeTypeLabel(type: NodeType): string {
    switch(type) {
        case NodeType.Http:
            return "HTTP"
        case NodeType.UsbCopy:
            return "USB COPY"
        case NodeType.Custom:
            return "CUSTOM"
        default:
            return "UNKNOWN TYPE"
    }
}
