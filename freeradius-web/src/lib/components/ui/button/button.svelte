<script lang="ts">
	import { Button as ButtonPrimitive } from 'bits-ui';
	import { buttonVariants, cn } from './index.js';
	import type { ButtonSize, ButtonVariant } from './index.js';

	export let variant: ButtonVariant = 'default';
	export let size: ButtonSize = 'default';
	export let href: string | undefined = undefined;
	export let type: 'button' | 'submit' | 'reset' = 'button';

	$: klass = cn(buttonVariants({ variant, size }), $$props.class);
</script>

{#if href}
	<ButtonPrimitive.Root {href} class={klass} {...$$restProps} on:click on:keydown>
		<slot />
	</ButtonPrimitive.Root>
{:else}
	<ButtonPrimitive.Root class={klass} {type} {...$$restProps} on:click on:keydown>
		<slot />
	</ButtonPrimitive.Root>
{/if}
