const RequestURL = document.getElementById("bundle").dataset.beUrl;
export const FrontendURL = window.location.href.split("/").slice(0,-1).join("/")
const CredentialsValue = FrontendURL.startsWith("http://localhost") ? "include" : "same-origin"
const SOCKET_ID_HEADER_KEY = "StackNoteProto-SID";
let defaultHeaders: [string, string][] = [];

// EXAMPLE
// export const reqTrashNotesById = (ids: uuid[]): Promise<Response> => {    
//     return new Promise((resolve, reject) => {
//         fetch(`${RequestURL}/note/trash`, {
//             method: 'POST',
//             credentials: CredentialsValue,
//             headers: defaultHeaders,
//             body: JSON.stringify({
//                 ids: ids
//             })
//         }).then(resp => {
//             resolve(handleResponse(resp));
//         }).catch(err => {
//             reject(err)
//         })
//     }); 
// }

const handleResponse = (r: Response): Promise<Response> => {
    return new Promise<Response>((res, rej) => {
        if (r.status === 200) {
            res(r);
            return;
        }
        r.text().then(err => {
            rej(err);
        });
    });
}