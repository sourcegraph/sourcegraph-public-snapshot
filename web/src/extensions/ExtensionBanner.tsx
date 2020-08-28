import React from 'react'

export const ExtensionBanner: React.FunctionComponent = React.memo(() => {
    const bannerPuzzleIcon = (
        <svg
            className="m-3 flex-shrink-0"
            width="69"
            height="62"
            viewBox="0 0 69 62"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
        >
            <rect x="53" y="24" width="6.00004" height="6.00004" fill="#00B4F2" />
            <rect x="53" y="38" width="6.00004" height="6.00004" fill="#F96216" />
            <rect x="63" y="31.9995" width="6.00004" height="6.00004" fill="#B200F8" />
            <path
                d="M46.9317 38.8646H41.6219V46.2875H34.6197V36.9134H28.2732V22.9582H38.6708V29.1723H46.9423C47.524 29.1723 47.9894 28.7057 48 28.1225V16.1821C48 15.9382 47.9788 15.7049 47.926 15.4822C47.799 14.7717 47.4606 14.1991 47.0057 13.7537C46.974 13.7219 46.9423 13.7007 46.9105 13.6689C46.7942 13.5628 46.6673 13.4568 46.5403 13.3614C46.4028 13.2659 46.2442 13.1811 46.0961 13.0962C46.0326 13.0644 45.9797 13.0326 45.9163 13.0008C45.8845 12.9902 45.8528 12.9796 45.8316 12.969C45.1653 12.6827 44.4354 12.5342 43.7585 12.5342L41.7805 12.5766C41.7805 12.5766 39.2737 12.6615 36.3966 12.7251H29.1723C28.58 12.6933 28.1781 12.6509 28.03 12.5766C26.4328 11.8343 29.9127 9.20449 30.4416 7.68808C31.8167 3.69028 28.2309 0 24 0C19.7691 0 16.1833 3.69028 17.569 7.69868C18.0978 9.21509 21.7576 11.8768 19.9806 12.5872C19.7479 12.6827 18.4152 12.7251 16.6699 12.7251H15.0304C11.0745 12.7145 6.2089 12.5978 6.2089 12.5978L4.23094 12.5554C2.46452 12.5554 0.401939 13.3826 0.0951961 15.3762C0.0423094 15.6413 0 16.1927 0 16.1927L0.031732 31.0356C0.0528867 31.2795 0.0740414 31.4597 0.105773 31.5234C0.814456 33.3049 4.69634 28.6527 6.2089 28.1225C10.2071 26.7227 12.6294 32.7353 12.6294 36.9876C12.6294 41.2399 8.94843 44.8242 4.9502 43.435C3.43764 42.9048 0.825033 39.416 0.0740414 41.0172C0.0423094 41.0703 0.0211547 41.1763 0 41.3248V57.7826C0.031732 59.6701 1.55487 61.1865 3.43764 61.1865H19.3671C19.5787 61.1653 19.7373 61.1441 19.8008 61.1229C21.5778 60.4124 19.1385 58.7402 18.6626 57.0541C17.499 52.9715 19.7585 52.2551 24 52.2551C28.2415 52.2551 30.2512 53.3216 30.2512 56.2343C30.2512 57.7295 26.2424 60.37 27.8396 61.1229C27.8925 61.1441 27.9877 61.1653 28.1146 61.1865H44.5306C46.4346 61.1865 47.9788 59.6383 47.9788 57.7295V39.9144C47.9788 39.3312 47.5134 38.8646 46.9317 38.8646Z"
                fill="url(#paint0_linear)"
                fillOpacity="0.32"
            />
            <defs>
                <linearGradient id="paint0_linear" x1="24" y1="0" x2="24" y2="61.1865" gradientUnits="userSpaceOnUse">
                    <stop stopColor="#95A5C6" />
                    <stop offset="1" stopColor="#95A5C6" stopOpacity="0.6" />
                </linearGradient>
            </defs>
        </svg>
    )

    return (
        <>
            <hr className="extension-banner__divider" />

            <div className="extension-banner">
                <div className="extension-banner__card">
                    {bannerPuzzleIcon}
                    <div className="extension-banner__text-container">
                        <h3>Create your own extension</h3>
                        <p>
                            You can improve your workflow by creating custom extensions. Read the Sourcegraph Docs for
                            details about writing and publishing.
                        </p>
                        <a className="btn btn-primary mt-2" href="https://docs.sourcegraph.com/extensions/authoring">
                            Learn more
                        </a>
                    </div>
                </div>
            </div>
        </>
    )
})
