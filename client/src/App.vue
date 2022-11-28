<script setup>
import { ref } from 'vue'
import { getEmails } from './services/indexer';

const search = ref('')
const emails = ref([])
const loading = ref(false)
const query = ref('')
const messageBody = ref('')
const messageSubject = ref('')
const isMessageVisible = ref(false)

const handleArrow = () => {
    isMessageVisible.value = !isMessageVisible.value
}

const handleClickEmail = (email) => {
    messageBody.value = email._source.body
    messageSubject.value = email._source.subject
    isMessageVisible.value = true
}

const searchEmails = async (query) => {
    loading.value = true
    emails.value = await getEmails(query)
    loading.value = false
}

function handleSearch() {
    isMessageVisible.value = false
    query.value = `search=${search.value}&page=1&page-size=100`
    console.log(search.value)
    searchEmails(query.value)
}

</script>

<template>
    <div class="grid min-h-full max-h-full h-full grid-rows-[auto_1fr]">
        <header class="flex justify-center pb-2 sm:pb-4">
            <svg xmlns="http://www.w3.org/2000/svg" xml:space="preserve"
                style="enable-background:new 0 0 392.533 392.533" viewBox="0 0 392.5 392.5" class="h-8 lg:h-10">
                <path
                    d="M22 230c0 6 5 11 11 11h126a85 85 0 0 1 169 0h12c6 0 10-5 10-11V37l-131 98a57 57 0 0 1-65 0L22 37"
                    style="fill:#fff" />
                <circle cx="243.7" cy="243.6" r="63.2" style="fill:#fff" />
                <circle cx="243.7" cy="243.6" r="40.8" style="fill:#ffc10d" />
                <path
                    d="M371 348c0-11-23-34-56-58-7 10-15 18-24 24 23 34 46 57 57 57 6-1 22-17 23-23zM206 117l128-95H38l128 95c12 8 28 8 40 0z"
                    style="fill:#56ace0" />
                <path
                    d="m324 271 2-9h14c18 0 32-14 32-32V33c0-18-14-33-32-33H33C15 0 0 15 0 33v197c0 18 15 33 33 33h128a85 85 0 0 0 110 61c22 33 52 69 77 69 18 0 44-27 44-45 1-25-35-55-68-77zm-34 43c10-6 18-14 24-24 34 24 57 47 57 58-1 6-17 22-23 23-11 0-34-23-58-57zM22 230v-29h24a11 11 0 1 0 0-22H22v-22h56a11 11 0 1 0 0-22H22V37l132 98c20 14 45 13 65 0l131-98v193c0 6-5 11-11 11h-11a85 85 0 0 0-169 0H33c-6 0-11-5-11-11zm144-113L38 22h296l-128 95c-12 8-28 8-40 0zm14 127a63 63 0 1 1 127 0 63 63 0 0 1-127 0z"
                    style="fill:#194f82" />
            </svg>
            <h1 class="pl-4 text-lg font-semibold sm:text-2xl lg:text-3xl">Enron Email</h1>
        </header>

        <main class="flex flex-col h-full overflow-hidden">
            <div class="w-full max-w-5xl m-auto">
                <form @submit.prevent="handleSearch">
                    <label for="default-search"
                        class="mb-2 text-sm font-medium text-gray-900 sr-only dark:text-white">Search</label>
                    <div class="relative">
                        <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                            <svg aria-hidden="true" class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none"
                                stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                            </svg>
                        </div>
                        <input type="search" id="default-search" v-model="search"
                            class="block w-full p-4 pl-10 text-sm text-gray-900 border border-gray-300 rounded-lg bg-gray-50 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                            placeholder="Search Message, From, To and Date" required>
                        <button type="submit" class="text-white absolute right-2.5 bottom-2.5
                            bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none
                            focus:ring-blue-300 font-medium rounded-lg text-sm px-4 py-2
                            dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
                            Search
                        </button>
                    </div>
                </form>
            </div>

            <div class="grow flex flex-col m-0 p-0 mt-4 h-full overflow-hidden pb-1 lg:flex-row">
                <div class="grow h-full max-h-full overflow-scroll shadow-md sm:rounded-lg">
                    <table class="w-[639px] table-fixed text-sm text-left text-gray-500 sm:w-full">
                        <thead class="sticky top-0 leading-8 text-xs uppercase bg-gray-100">
                            <tr class="">
                                <th scope="col" class=" py-3 px-5 font-semibold">
                                    Subject
                                </th>
                                <th scope="col" class=" py-3 px-5 font-semibold">
                                    From
                                </th>
                                <th scope="col" class=" py-3 px-5 font-semibold">
                                    To
                                </th>
                                <th scope="col" class=" py-3 px-5 font-semibold">
                                    Date
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr class="bg-white border-b break-words cursor-pointer" v-for="email in emails"
                                :key="email._id" @click="handleClickEmail(email)">
                                <td class="p-2 lg:py-3">
                                    {{ email._source.subject }}
                                </td>
                                <td class="p-2 lg:py-3">
                                    {{ email._source.from }}
                                </td>
                                <td class="p-2 lg:py-3">
                                    {{ email._source.to }}
                                </td>
                                <td class="p-2 lg:py-3">
                                    {{ email._source.date }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="flex relative text-sm transition-all duration-500 ease-in-out lg:w-2/5 lg:min-w-[40%] lg:h-full lg:border-0 lg:mx-5 lg:p-0"
                    :class="[isMessageVisible ? 'h-60 mt-4 p-2 border-t-2 lg:my-0' : 'h-0 p-0 border-0']">
                    <div class="absolute -top-7 right-2 cursor-pointer px-2 py-1 transition-all duration-300 ease-in-out
                    bg-gray-100 hover:bg-gray-200 rounded lg:hidden" @click="handleArrow"
                        :class="[isMessageVisible ? '' : 'rotate-180 -top-7']">
                        <svg xmlns="http://www.w3.org/2000/svg" xml:space="preserve" class="h-5 fill-gray-700"
                            :class="[]" style="enable-background:new 0 0 330 330" viewBox="0 0 330 330">
                            <path
                                d="M325.607 79.393c-5.857-5.857-15.355-5.858-21.213.001l-139.39 139.393L25.607 79.393c-5.857-5.857-15.355-5.858-21.213.001-5.858 5.858-5.858 15.355 0 21.213l150.004 150a14.999 14.999 0 0 0 21.212-.001l149.996-150c5.859-5.857 5.859-15.355.001-21.213z" />
                        </svg>
                    </div>
                    <div class="w-full h-full overflow-scroll">
                        <h3 class="font-semibold pb-1 sticky top-0 bg-white lg:pb-4">
                            {{ messageSubject }}
                        </h3>
                        {{ messageBody }}
                    </div>
                </div>
            </div>
        </main>
    </div>


</template>