/**
 * Returns emails containing the given search term.
 * 
 * @param {string} query - url query string
 *
 */

export async function getEmails(query) {
    const API_URL = import.meta.env.VITE_API_URL;
    console.log(API_URL);
    try {
        const response = await fetch(`${API_URL}/emails?${query}`);
        if (!response.ok) {
            throw new Error(response.statusText);
        }
        const data = await response.json();
        // window.history.pushState({}, '', `/?${query}`);
        return data;
    } catch (error) {
        console.error(error);
    }
}