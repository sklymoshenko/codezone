import { SiGo, SiJavascript, SiPostgresql } from 'solid-icons/si'
import { Component } from 'solid-js'
import type { Language } from '../types'
import LangButton from './LangButton'

type LangSwitchProps = {
  currentLanguage: Language
  onLanguageChange: (lang: Language) => void
}

const LangSwitch: Component<LangSwitchProps> = props => {
  return (
    <div class="flex space-x-2">
      <LangButton
        onClick={() => props.onLanguageChange('javascript')}
        isActive={props.currentLanguage === 'javascript'}
        icon={<SiJavascript size={16} />}
        activeClasses="bg-yellow-400 text-black"
        hoverClasses="hover:bg-yellow-400 hover:text-black"
      />
      <LangButton
        onClick={() => props.onLanguageChange('go')}
        isActive={props.currentLanguage === 'go'}
        icon={<SiGo size={16} />}
        activeClasses="bg-cyan-400 text-white"
        hoverClasses="hover:bg-cyan-400"
      />
      <LangButton
        onClick={() => props.onLanguageChange('postgres')}
        isActive={props.currentLanguage === 'postgres'}
        icon={<SiPostgresql size={16} />}
        activeClasses="bg-blue-600 text-white"
        hoverClasses="hover:bg-blue-600"
      />
    </div>
  )
}

export default LangSwitch
