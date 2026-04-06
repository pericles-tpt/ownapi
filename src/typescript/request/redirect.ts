import { FrontendURL } from "./request"

export function redirectToHome() {
    window.location.replace(`${FrontendURL}`)
}