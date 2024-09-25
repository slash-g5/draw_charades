class CssClassGenerator {
    constructor(color) {
        this.indvScoreBaseClass = `flex flex-row border-2 border-` + theme + `-400` +  ` bg-` + theme + `-200 mt-4 lg:mt-8 rounded-lg lg:p-2 truncate w-52 lg:w-80`;
    }
}

function loadTheme(root, theme) {
      // Define the current and new theme patterns to replace
  const themeClasses = {
    bg: "bg-",
    hoverBg: "hover:bg-",
    border: "border-",
    text: "text-"
  };

  const avoidClasses = ["border-gray-800",
    "bg-white",
    "text-black",
    "text-gray-800",
    "border-black", 
    "border-2",
    "text-center",
    "text-xl",
    "text-2xl",
    "text-3xl"
  ]

  // Define the new theme class templates
  const newClasses = {
    bg: `bg-${theme}-100`,
    hoverBg: `hover:bg-${theme}-400`,
    border: `border-${theme}-400`,
    text: `text-${theme}-500`
  };

  const toAdd = [];
  const toRemove = [];

  root.classList.forEach(className => {
    if(className.startsWith(themeClasses.bg) && !avoidClasses.includes(className)) {
        toRemove.push(className)
        toAdd.push(newClasses.bg);
    }
    else if(className.startsWith(themeClasses.border) && !avoidClasses.includes(className)) {
        toRemove.push(className);
        toAdd.push(newClasses.border);
    }
    else if(className.startsWith(themeClasses.hoverBg)) {
        toRemove.push(className);
        toAdd.push(newClasses.hoverBg);
    }
    else if(className.startsWith(themeClasses.text) && !avoidClasses.includes(className)) {
        toRemove.push(className);
        toAdd.push(newClasses.text)
    }
  });

  for(const className of toRemove) {
    root.classList.remove(className);
  }

  for(const className of toAdd) {
    root.classList.add(className);
  }
}

function loadThemeRecursive(root, theme) {

    loadTheme(root, theme)

    const elements = root.querySelectorAll("*");
    
    for(const element of elements) {
        loadTheme(element, theme);
    }
}

export {loadTheme, loadThemeRecursive}